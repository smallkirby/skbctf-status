package main

import (
	"io/ioutil"
	"path/filepath"

	"go.uber.org/zap"
)

func CheckAll(logger zap.SugaredLogger, dir string) error {
	challs_dir, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range challs_dir {
		if f.IsDir() {
			res := make(chan TestResult, 1)
			abspath, _ := filepath.Abs(filepath.Join(dir, f.Name()))
			executer := Executer{path: abspath, logger: logger}
			go executer.check(res)
		}
	}

	return nil
}
