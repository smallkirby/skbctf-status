package main

import (
	"flag"
	"log"
	"time"

	"github.com/smallkirby/skbctf-status/checker"
	"go.uber.org/zap"
)

func create_conf(logger zap.SugaredLogger) checker.CheckerConfig {
	// commandline option overrides configuration from config file.
	// priority of config is: command-line > config file > default
	conffile := flag.String("config", "checker.example.conf.json", "Config file name of checker.")
	timeout := flag.Float64("timeout", 10.0, "Timeout for solve checks.")
	single := flag.Bool("single", false, "Execute test set only once and exit.")
	parallel := flag.Bool("parallel", true, "Execute solve checks in parallel.")
	infofile := flag.String("infofile", "info.json", "File name of configuration file for each challs.")
	nodb := flag.Bool("nodb", false, "Not write to DB.")
	challs_dir := flag.String("challs", "examples", "Challenges directory path.")
	interval := flag.Int("interval", 30, "Testing interval in minutes.")
	flag.Parse()

	// create default config
	conf, err := checker.ReadConf(*conffile)
	if err != nil {
		logger.Infof("failed to read config file:\n%w", err)
		logger.Info("Defaulting to default values...")

		// use default values
		conf.Nodb = *nodb
		conf.Parallel = *parallel
		conf.Single = *single
		conf.Timeout = *timeout
		conf.Infofile = *infofile
		conf.ChallsDir = *challs_dir
		conf.Interval = *interval
	}

	// Overwrite with command-line options
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
		case "challs":
			conf.ChallsDir = *challs_dir
		case "interval":
			conf.Interval = *interval
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
		log.Fatalf("[!] failed to init logger:\n%s", err)
	}
	slogger := logger.Sugar()

	conf := create_conf(*slogger)

	for {
		if err := checker.CheckAllOnce(*slogger, conf); err != nil {
			slogger.Warnf("Fatal error detected:\n%v", err)
		}

		if conf.Single {
			break
		} else {
			time.Sleep(time.Duration(conf.Interval) * time.Minute)
		}
	}
}
