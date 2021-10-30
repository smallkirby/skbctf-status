package checker

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type CheckerConfig struct {
	Single    bool    `json:"single"`
	Parallel  bool    `json:"parallel"`
	Timeout   float64 `json:"timeout"`
	Infofile  string  `json:"infofile"`
	Nodb      bool    `json:"nodb"`
	ChallsDir string  `json:"challs"`
	Interval  int     `json:"interval"`
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
