package checker

/***
* This file defines fields of config for checker.
* Also, it implements helper function of config.
***/

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type CheckerConfig struct {
	Single      bool    `json:"single"`
	Parallel    bool    `json:"parallel"`
	ParallelNum uint    `json:"pnum"`
	Timeout     float64 `json:"timeout"`
	Infofile    string  `json:"infofile"`
	Nodb        bool    `json:"nodb"`
	ChallsDir   string  `json:"challs"`
	Interval    uint    `json:"interval"`
	Retries     uint    `json:"retries"`
}

func (ch *CheckerConfig) ResolveConflict() {
	if ch.ParallelNum >= 1 {
		ch.Parallel = true
	}
	if ch.Parallel && ch.ParallelNum <= 0 {
		ch.ParallelNum = 1
	}
	if ch.Interval == 0 {
		ch.Interval = 1
	}
}

func (ch *CheckerConfig) Validate() {
	if ch.Timeout < 0 {
		ch.Timeout = 0
	}
}

func ReadConf(filename string) (CheckerConfig, error) {
	if filepath.IsAbs(filename) {
		cwd, _ := os.Getwd()
		filename = filepath.Join(cwd, filename)
	}

	conf_file, err := os.Open(filename)
	if err != nil {
		return CheckerConfig{}, err
	}
	conf_bytes, err := ioutil.ReadAll(conf_file)
	if err != nil {
		return CheckerConfig{}, err
	}

	var conf CheckerConfig
	if err := json.Unmarshal(conf_bytes, &conf); err != nil {
		return conf, err
	}

	return conf, nil
}
