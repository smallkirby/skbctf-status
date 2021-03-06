package checker

/***
* This file implements an actual executer of tests.
* It doesn't record to DB.
***/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"
)

/***
* Executer of test.
* @path: path to challenge directory
* @retry_max: maximum # of execution retry
* @try_current: # of executed try
***/
type Executer struct {
	path        string
	logger      zap.SugaredLogger
	retry_max   uint
	try_current uint
}

/***
* Information of challenge.
* Used to hold information to execute test.
* @Name: challenge name
* @Id: challenge ID
* @Default_success: if test is not executed due to insufficient info, result becomes `Success` if this is true.
* @Result: test result
***/
type Challenge struct {
	Name             string  `json:"name"`
	Id               int     `json:"id"`
	Default_success  bool    `json:"default"`
	Timeout          float64 `json:"timeout"`
	Exploit_dir_name string
	Result           TestResult
}

/***
* Test result.
***/
type TestResult int

const (
	// Test is successful
	TestSuccess TestResult = iota
	// Test is not executed due to insufficient info, but defaults to Success
	TestSuccessWithoutExecution
	// Test didn't end before timeout
	TestTimeout
	// Test is not executed due to insufficient info, and falls into Failure
	TestNotExecuted
	// Test is failure
	TestFailure
)

func (tr TestResult) String() string {
	switch tr {
	case TestSuccess:
		return "TestSuccess"
	case TestTimeout:
		return "TestTimeout"
	case TestNotExecuted:
		return "TestNotExecuted"
	case TestFailure:
		return "TestFailure"
	case TestSuccessWithoutExecution:
		return "TestSuccessWithoutExecution"
	default:
		return "UnknownFailure"
	}
}

func (tr TestResult) ToMessage() string {
	switch tr {
	case TestSuccess, TestSuccessWithoutExecution:
		return "Success"
	case TestTimeout:
		return "Timeout"
	case TestNotExecuted:
		return "test not executed"
	case TestFailure:
		return "Failure"
	default:
		return "UnknownFailure"
	}
}

func (tr TestResult) ToColor() string {
	switch tr {
	case TestSuccess:
		return "33FF99"
	case TestSuccessWithoutExecution:
		return "66FF66"
	case TestTimeout:
		return "6600CC"
	case TestNotExecuted:
		return "808080"
	case TestFailure:
		return "CC0000"
	default:
		return "202020"
	}
}

/***
* Check challenge directory and collect challenge information to check whether test can be executed.
***/
func (e *Executer) prepare_check(infofile string) (Challenge, error) {
	ret := TestNotExecuted

	// read config file
	cfg_file_name := filepath.Join(e.path, infofile)
	cfg_file, err := os.Open(cfg_file_name)
	if err != nil {
		return Challenge{Result: ret}, fmt.Errorf("%s not found in %s.", infofile, e.path)
	}
	cfg_bytes, err := ioutil.ReadAll(cfg_file)
	if err != nil {
		return Challenge{Result: ret}, fmt.Errorf("Failed to read config file %s.", cfg_file_name)
	}

	var chall Challenge
	if err := json.Unmarshal(cfg_bytes, &chall); err != nil {
		return Challenge{Result: ret}, fmt.Errorf("Failed to parse %s as JSON:\n%v", cfg_file_name, err)
	}
	if chall.Default_success {
		ret = TestSuccessWithoutExecution
	}

	// check exploit path and Dockerfile
	chall.Exploit_dir_name = filepath.Join(e.path, "exploit")
	if _, err := os.Stat(chall.Exploit_dir_name); os.IsNotExist(err) {
		// check whether this is symlink
		realpath, err := os.Readlink(chall.Exploit_dir_name)
		if err != nil {
			chall.Result = ret
			return chall, fmt.Errorf("[%s] Exploit dir not found. (failed to read symlink of dir)", chall.Name)
		}
		realpath = filepath.Join(e.path, realpath)
		e.logger.Infof("symlink found: %s -> %s", chall.Exploit_dir_name, realpath)
		chall.Exploit_dir_name = realpath
	}

	docker_file_name := filepath.Join(chall.Exploit_dir_name, "Dockerfile")
	if _, err := os.Stat(docker_file_name); os.IsNotExist(err) {
		chall.Result = ret
		return chall, fmt.Errorf("[%s] Dockerfile not found in exploit dir: %s", chall.Name, chall.Exploit_dir_name)
	}

	return chall, nil
}

/***
* Do execute test and return result via `res_chan` channel.
* This function receives channel for kill signal for timeout.
***/
func (e *Executer) execute_internal(res_chan chan Challenge, chall Challenge, killer_chan <-chan bool) {
	// prepare command
	container_name := fmt.Sprintf("container_solver_%d_%d", chall.Id, time.Now().Unix())
	image_name := fmt.Sprintf("solver_%d", chall.Id)
	cmd := exec.Command("bash", "-c", fmt.Sprintf("docker run --name %s --rm $(docker build -qt solver_%s %s)", container_name, image_name, chall.Exploit_dir_name))

	// termination signal hook
	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan, os.Interrupt, syscall.SIGTERM)

	// execute test async
	if err := cmd.Start(); err != nil {
		e.logger.Warnf("[%s] Failed to start test: \n%w", chall.Name, err)
		chall.Result = TestFailure
		res_chan <- chall
		return
	}
	e.logger.Infof("[%s] Test started as pid %d in %s.", chall.Name, cmd.Process.Pid, container_name)

	res_chan_internal := make(chan error)
	var errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	go func() {
		res_chan_internal <- cmd.Wait()
	}()

	shutdown_hook := func() {
		// remove container
		if err := exec.Command("docker", "rm", "-f", container_name).Run(); err != nil {
			e.logger.Warnf("Failed to remove container(%s):\n%v", container_name, err)
		}
		// kill process
		if err := cmd.Process.Kill(); err != nil {
			e.logger.Warnf("Failed to kill process: %v", err)
		}
		res_chan <- chall
	}

	// wait and get exit-status
	select {
	case <-signal_chan: // process terminated
		e.logger.Info("Process terminated, cleaning up docker container...")
		shutdown_hook()
		os.Exit(0)
	case _, ok := <-killer_chan: // timeout
		if !ok {
			shutdown_hook()
		}
		break
	case err := <-res_chan_internal: // test execution end
		if err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					exit_code := status.ExitStatus()
					e.logger.Infof("[%s] Test failed with status %d.\n%v", chall.Name, exit_code, err)
					e.logger.Infof("%v", errbuf.String())
					chall.Result = TestFailure
					res_chan <- chall
					return
				}
			}
		} else {
			// command ends without any failure
			e.logger.Infof("[%s] exits with status code 0.", chall.Name)
			chall.Result = TestSuccess
			res_chan <- chall
		}
	}
}

/***
*	execute tests w/o timeout.
*	it retries execution for specified times if a test fails.
***/
func (e *Executer) Check(res_chan chan<- Challenge, infofile string) {
	// read config file and check target directry structure
	chall, err := e.prepare_check(infofile)
	if err != nil {
		e.logger.Infof("%v", err)
		res_chan <- chall
		return
	}

	// execute test
	for e.try_current <= e.retry_max {
		res_chan_internal := make(chan Challenge)
		killer_chan := make(chan bool)
		go e.execute_internal(res_chan_internal, chall, killer_chan)
		chall = <-res_chan_internal

		// retry a test or return result
		if chall.Result == TestSuccess || chall.Result == TestSuccessWithoutExecution {
			break
		}
		if e.try_current < e.retry_max {
			e.logger.Infof("[%s] Retrying test...", chall.Name)
		}
		e.try_current++
	}

	res_chan <- chall
}

/***
*	execute tests with timeout.
*	it retries execution for specified times if a test fails.
***/
func (e *Executer) CheckWithTimeout(res_chan chan<- Challenge, infofile string, timeout float64) {
	// read config file and check target directry structure
	chall, err := e.prepare_check(infofile)
	if err != nil {
		e.logger.Infof("%v", err)
		res_chan <- chall
		return
	}

	// overwrite timeout if specified for this chall
	if chall.Timeout > 0 {
		timeout = chall.Timeout
	}

	e.logger.Infof("[%s] timeout set to %f.", chall.Name, timeout)

	// execute test some times
	for e.try_current <= e.retry_max {
		res_chan_internal := make(chan Challenge)
		killer_chan := make(chan bool)
		go e.execute_internal(res_chan_internal, chall, killer_chan)

		// wait end of execution, or kill process for timeout.
		select {
		case result := <-res_chan_internal:
			chall = result
			break
		case <-time.After(time.Duration(timeout) * time.Second):
			close(killer_chan)
			chall = <-res_chan_internal
			chall.Result = TestTimeout
			break
		}

		// retry a test or return result
		if chall.Result == TestSuccess || chall.Result == TestSuccessWithoutExecution {
			break
		}
		if e.try_current < e.retry_max {
			e.logger.Infof("[%s] Retrying test...", chall.Name)
		}
		e.try_current++
	}

	res_chan <- chall
}
