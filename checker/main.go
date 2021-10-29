package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"go.uber.org/zap"
)

type CheckerConfig struct {
	Single   bool    `json:"single"`
	Parallel bool    `json:"parallel"`
	Timeout  float64 `json:"timeout"`
	Infofile string  `json:"info.json"`
	Nodb     bool    `json:"nodb"`
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

func create_conf(logger zap.SugaredLogger) CheckerConfig {
	// commandline option overrides configuration from config file.
	// priority of config is: command-line > config file > default
	conffile := flag.String("config", "checker.conf.json", "Config file name of checker.")
	timeout := flag.Float64("timeout", 10.0, "Timeout for solve checks.")
	single := flag.Bool("single", false, "Execute test set only once and exit.")
	parallel := flag.Bool("parallel", true, "Execute solve checks in parallel.")
	infofile := flag.String("infofile", "info.json", "File name of configuration file for each challs.")
	nodb := flag.Bool("nodb", false, "Not write to DB.")
	flag.Parse()

	// create default config
	conf, err := read_conf(*conffile)
	if err != nil {
		logger.Infof("failed to read config file:\n%w", err)
		logger.Info("Defaulting to default values...")

		// use default values
		conf.Nodb = *nodb
		conf.Parallel = *parallel
		conf.Single = *single
		conf.Timeout = *timeout
		conf.Infofile = *infofile
	}
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "nodb":
			conf.Nodb = *nodb
		case "timeout":
			conf.Timeout = *timeout
		case "single":
			conf.Single = *single
		case "parallel":
			conf.Parallel = *parallel
		case "config":
			break
		default:
			logger.Errorf("Unknown flag found: %s", f.Name)
		}
	})

	return conf
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("[!] failed to init logger:\n%w", err)
	}
	slogger := logger.Sugar()

	conf := create_conf(*slogger)
	fmt.Print(conf)
}
