package main

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

func TestBlockingExecute(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	slogger := logger.Sugar()

	challs_dir, err := ioutil.ReadDir("../examples")
	if err != nil {
		t.Errorf("Challenges dir not found: %v.", challs_dir)
		return
	}

	for _, f := range challs_dir {
		if f.IsDir() {
			res := make(chan TestResult, 1)
			abspath, _ := filepath.Abs(filepath.Join("../examples", f.Name()))
			executer := Executer{path: abspath, logger: *slogger}
			executer.check(res)
			result := <-res
			slogger.Infof("Result: %v", result)
			switch result {
			case TestSuccess:
			case TestSuccessWithoutExecution:
				continue
			default:
				t.Error("Test failed.")
			}
		}
	}
}
