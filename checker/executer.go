package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"go.uber.org/zap"
)

type Executer struct {
	path   string
	logger zap.SugaredLogger
}

type Challenge struct {
	Name            string `json:"name"`
	Id              int    `json:"id"`
	Default_success bool   `json:"default"`
	result          TestResult
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

func (e Executer) check(res_chan chan<- Challenge, infofile string) {
	ret := TestNotExecuted

	// read config file
	cfg_file_name := filepath.Join(e.path, infofile)
	cfg_file, err := os.Open(cfg_file_name)
	if err != nil {
		e.logger.Infof("%s not found in %s.", infofile, e.path)
		res_chan <- Challenge{result: ret}
		return
	}
	cfg_bytes, err := ioutil.ReadAll(cfg_file)
	if err != nil {
		e.logger.Infof("Failed to read config file %s.", cfg_file_name)
		res_chan <- Challenge{result: ret}
		return
	}

	var chall Challenge
	if err := json.Unmarshal(cfg_bytes, &chall); err != nil {
		e.logger.Warnf("Failed to parse %s as JSON.", cfg_file_name)
		e.logger.Warnf("%w", err)
		res_chan <- Challenge{result: ret}
		return
	}
	if chall.Default_success {
		ret = TestSuccessWithoutExecution
	}

	// check exploit path and Dockerfile
	exploit_dir_name := filepath.Join(e.path, "exploit")
	if _, err := os.Stat(exploit_dir_name); os.IsNotExist(err) {
		e.logger.Infof("[%s] Exploit dir not found.", chall.Name)
		chall.result = ret
		res_chan <- chall
		return
	}
	docker_file_name := filepath.Join(exploit_dir_name, "Dockerfile")
	if _, err := os.Stat(docker_file_name); os.IsNotExist(err) {
		e.logger.Infof("[%s] Dockerfile not found in exploit dir.", chall.Name)
		chall.result = ret
		res_chan <- chall
		return
	}

	// execute test
	cmd := exec.Command("bash", "-c", fmt.Sprintf("docker run $(docker build -qt solver_%d %s)", chall.Id, exploit_dir_name))

	if err = cmd.Start(); err != nil {
		e.logger.Warnf("[%s] Failed to start test: \n%w", chall.Name, err)
		chall.result = TestFailure
		res_chan <- chall
		return
	}
	e.logger.Infof("[%s] Test started as pid: %d.", chall.Name, cmd.Process.Pid)

	// wait and get exit-status
	if err = cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exit_code := status.ExitStatus()
				e.logger.Infof("[%s] Test failed with status %d.", chall.Name, exit_code)
				chall.result = TestFailure
				res_chan <- chall
				return
			}
		}
	}

	// command ends without any failure
	e.logger.Infof("[%s] exits with status code 0.", chall.Name)
	chall.result = TestSuccess
	res_chan <- chall
}
