package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type CheckerConfig struct {
	Single    bool    `json:"single"`
	Parallel  bool    `json:"parallel"`
	Timeout   float64 `json:"timeout"`
	Infofile  string  `json:"info.json"`
	Nodb      bool    `json:"nodb"`
	ChallsDir string  `json:"challs"`
}

func read_conf(filename string) (CheckerConfig, error) {
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
