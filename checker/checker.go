package main

import (
	"io/ioutil"
	"path/filepath"

	"go.uber.org/zap"
)

func CheckAll(logger zap.SugaredLogger, conf CheckerConfig) error {
	// enumerate challenge dirs
	challs_dir, err := ioutil.ReadDir(conf.ChallsDir)
	if err != nil {
		return err
	}

	// execute tests
	if conf.Parallel {
		logger.Error("not implemented")
	} else {
		for _, f := range challs_dir {
			if f.IsDir() {
				ch := make(chan Challenge, 1)
				abspath, _ := filepath.Abs(filepath.Join(conf.ChallsDir, f.Name()))
				executer := Executer{path: abspath, logger: logger}
				go executer.check(ch, conf.Infofile)

				chall := <-ch
				logger.Infof("[%s] Test finish.", chall.Name)
			}
		}
	}

	return nil
}
