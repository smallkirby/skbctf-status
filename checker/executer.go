package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

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
	default:
		return "UnknownFailure"
	}
}

func (e Executer) check() TestResult {
	cfg_file_name := filepath.Join(e.path, "info.json")
	cfg_file, err := os.Open(cfg_file_name)
	if err != nil {
		e.logger.Infof("info.json not found in %s.", e.path)
		return TestNotExecuted
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
	}

	return TestSuccess
}
