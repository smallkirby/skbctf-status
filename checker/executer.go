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

type Config struct {
	Name            string `json:"name"`
	Id              int    `json:"id"`
	Default_success bool   `json:"default"`
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

// XXX must be async, callback or channel
func (e Executer) check() TestResult {
	ret := TestNotExecuted

	// read config file
	cfg_file_name := filepath.Join(e.path, "info.json")
	cfg_file, err := os.Open(cfg_file_name)
	if err != nil {
		e.logger.Infof("info.json not found in %s.", e.path)
		return ret
	}
	cfg_bytes, err := ioutil.ReadAll(cfg_file)
	if err != nil {
		e.logger.Infof("Failed to read from %s.", cfg_file_name)
		return TestNotExecuted
	}

	var cfg Config
	if err := json.Unmarshal(cfg_bytes, &cfg); err != nil {
		e.logger.Warnf("Failed to parse %s as JSON.", cfg_file_name)
		e.logger.Warnf("\t%w", err)
		return ret
	}
	if cfg.Default_success {
		ret = TestSuccess
	}

	// check exploit path and Dockerfile
	exploit_dir_name := filepath.Join(e.path, "exploit")
	if _, err := os.Stat(exploit_dir_name); os.IsNotExist(err) {
		e.logger.Infof("Exploit dir not found for %s.", cfg.Name)
		return ret
	}
	docker_file_name := filepath.Join(exploit_dir_name, "Dockerfile")
	if _, err := os.Stat(docker_file_name); os.IsNotExist(err) {
		e.logger.Infof("Dockerfile not found in exploit dir of %s.", cfg.Name)
		return ret
	}

	// execute test
	cmd := exec.Command("bash", "-c", fmt.Sprintf("docker run $(docker build -qt solver_%d %s)", cfg.Id, exploit_dir_name))

	if err = cmd.Start(); err != nil {
		e.logger.Warnf("Failed to start test of '%s': %w", cfg.Name, err)
		return TestFailure
	}
	e.logger.Infof("Test of '%s' started as pid: %d.", cfg.Name, cmd.Process.Pid)

	// wait and get exit-status
	if err = cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exit_code := status.ExitStatus()
				e.logger.Infof("Test of '%s' failed with status %d.", cfg.Name, exit_code)
				return TestFailure
			}
		}
	}

	// command ends without any failure
	e.logger.Infof("'%s' ends with status code 0.", cfg.Name)
	return TestSuccess
}
