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
			res := make(chan Challenge, 1)
			abspath, _ := filepath.Abs(filepath.Join("../examples", f.Name()))
			executer := Executer{path: abspath, logger: *slogger}
			executer.check(res, "info.json")
			chall := <-res
			slogger.Infof("Result: %v", chall.result)

			// tests which must fail
			if f.Name() == "chall2" {
				if chall.result != TestFailure {
					t.Errorf("Test(%s) which must fail succeeded: %v", f.Name(), chall)
				} else {
					continue
				}
			}

			// tests which must succeed
			switch chall.result {
			case TestSuccess:
			case TestSuccessWithoutExecution:
				continue
			default:
				t.Errorf("Test '%s' failed.", chall.Name)
			}
		}
	}
}

func TestBlockingAllCheck(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	slogger := logger.Sugar()
	conf := CheckerConfig{
		Single:    true,
		Parallel:  false,
		Timeout:   10.0,
		Infofile:  "info.json",
		Nodb:      true,
		ChallsDir: "../examples",
	}
	if err := CheckAll(*slogger, conf); err != nil {
		t.Errorf("Test failed: %v", err)
	}
}
