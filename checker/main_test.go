package main

import (
	"fmt"
	"io/ioutil"
	"testing"

	"go.uber.org/zap"
)

func TestExecute(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	slogger := logger.Sugar()

	challs_dir, err := ioutil.ReadDir("../examples")
	if err != nil {
		t.Errorf("Challenges dir not found: %v.", challs_dir)
		return
	}

	for _, f := range challs_dir {
		if f.IsDir() {
			executer := Executer{path: fmt.Sprintf("../examples/%s", f.Name()), logger: *slogger}
			result := executer.check()
			slogger.Infof("Result: %v", result)
		}
	}
}
