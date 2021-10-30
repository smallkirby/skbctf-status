package checker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type Executer struct {
	path   string
	logger zap.SugaredLogger
}

type Challenge struct {
	Name             string `json:"name"`
	Id               int    `json:"id"`
	Default_success  bool   `json:"default"`
	Exploit_dir_name string
	Result           TestResult
}

type TestResult int

const (
	TestSuccess TestResult = iota
	TestSuccessWithoutExecution
	TestTimeout
	TestNotExecuted
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

func (e Executer) prepare_check(infofile string) (Challenge, error) {
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
		chall.Result = ret
		return chall, fmt.Errorf("[%s] Exploit dir not found.", chall.Name)
	}
	docker_file_name := filepath.Join(chall.Exploit_dir_name, "Dockerfile")
	if _, err := os.Stat(docker_file_name); os.IsNotExist(err) {
		chall.Result = ret
		return chall, fmt.Errorf("[%s] Dockerfile not found in exploit dir.", chall.Name)
	}

	return chall, nil
}

func (e Executer) execute_internal(res_chan chan<- Challenge, chall Challenge, kill_signal <-chan bool) {
	// execute test
	container_name := fmt.Sprintf("container_solver_%d", chall.Id)
	image_name := fmt.Sprintf("solver_%d", chall.Id)
	cmd := exec.Command("bash", "-c", fmt.Sprintf("docker run --name %s --rm $(docker build -qt solver_%s %s)", container_name, image_name, chall.Exploit_dir_name))

	if err := cmd.Start(); err != nil {
		e.logger.Warnf("[%s] Failed to start test: \n%w", chall.Name, err)
		chall.Result = TestFailure
		res_chan <- chall
		return
	}
	e.logger.Infof("[%s] Test started as pid: %d.", chall.Name, cmd.Process.Pid)

	res_chan_internal := make(chan error)
	var errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	go func() {
		res_chan_internal <- cmd.Wait()
	}()

	// wait and get exit-status
	select {
	case err := <-res_chan_internal:
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
	case <-kill_signal:
		// kill process
		if err := cmd.Process.Kill(); err != nil {
			e.logger.Warnf("Failed to kill timeouted process: %v", err)
		}
		// remove container
		if err := exec.Command("docker", "rm", "-f", container_name).Run(); err != nil {
			e.logger.Warnf("Failed to remove timeouted container(%s):\n%v", container_name, err)
		}
	}
}

func (e Executer) Check(res_chan chan<- Challenge, infofile string) {
	// read config file and check target directry structure
	chall, err := e.prepare_check(infofile)
	if err != nil {
		e.logger.Infof("%v", err)
		res_chan <- chall
		return
	}

	// execute test
	res_chan_internal := make(chan Challenge)
	signal_chan := make(chan bool)
	go e.execute_internal(res_chan_internal, chall, signal_chan)
	res_chan <- <-res_chan_internal
}

func (e Executer) CheckWithTimeout(res_chan chan<- Challenge, infofile string, timeout float64) {
	// read config file and check target directry structure
	chall, err := e.prepare_check(infofile)
	if err != nil {
		e.logger.Infof("%v", err)
		res_chan <- chall
		return
	}

	e.logger.Infof("[%s] timeout set to %f.", chall.Name, timeout)

	// execute test
	res_chan_internal := make(chan Challenge)
	signal_chan := make(chan bool)
	go e.execute_internal(res_chan_internal, chall, signal_chan)
	select {
	case result := <-res_chan_internal:
		res_chan <- result
		break
	case <-time.After(time.Duration(timeout) * time.Second):
		chall.Result = TestTimeout
		signal_chan <- true
		res_chan <- chall
		break
	}
}
