package checker

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func enumerateChallsDir(challs_dirname string) ([]string, error) {
	// open challs directory
	challs := make([]string, 0)
	challs_dir, err := ioutil.ReadDir(challs_dirname)
	if err != nil {
		return challs, err
	}

	// enumerate challs
	for _, f := range challs_dir {
		if f.IsDir() {
			chall, _ := filepath.Abs(filepath.Join(challs_dirname, f.Name()))
			challs = append(challs, chall)
		}
	}

	return challs, nil
}

func CheckAllOnce(logger zap.SugaredLogger, conf CheckerConfig) error {
	// prepare DB
	var db *sqlx.DB
	var err error
	if !conf.Nodb {
		db, err = Connect(os.Getenv("DBUSER"), os.Getenv("DBPASS"), os.Getenv("DBHOST"), os.Getenv("DBNAME"))
		if err != nil {
			return err
		}
	}

	// enumerate challenge dirs
	challs, err := enumerateChallsDir(conf.ChallsDir)
	if err != nil {
		return err
	}

	logger.Infof("found %d tests.", len(challs))

	// execute tests
	if conf.Parallel {
		// Parallel test execution
		executers_wait_que := make([]Executer, 0)
		num_running := 0
		ch := make(chan Challenge, len(challs))

		// init executer and push them into waiting que
		for _, challdir := range challs {
			executer := Executer{path: challdir, logger: logger, retry_max: conf.Retries, try_current: 0}
			executers_wait_que = append(executers_wait_que, executer)
		}

		// exec all
		// XXX --asnyc-num option would specify # of running tests in a time
		for len(executers_wait_que) > 0 {
			executer := executers_wait_que[0]
			executers_wait_que = executers_wait_que[1:]
			go executer.CheckWithTimeout(ch, conf.Infofile, conf.Timeout)
			num_running++
		}

		// wait and get results
		for result := range ch {
			logger.Infof("[%s] Test execution finish.", result.Name)
			num_running--

			// write result to DB
			if !conf.Nodb {
				if err := RecordResult(db, result); err != nil {
					logger.Warn("%v", err)
				}
			}

			// end of tests
			if num_running == 0 && len(executers_wait_que) == 0 {
				close(ch)
			}
		}
	} else {
		// Sequential test execution
		for _, challdir := range challs {
			ch := make(chan Challenge, 1)
			executer := Executer{path: challdir, logger: logger, retry_max: conf.Retries, try_current: 0}

			// blocking execution of test
			go executer.CheckWithTimeout(ch, conf.Infofile, conf.Timeout)
			chall := <-ch

			// write a result to DB
			if !conf.Nodb {
				if err := RecordResult(db, chall); err != nil {
					logger.Warn("%v", err)
				}
			}
			logger.Infof("[%s] Test execution finish: %v", chall.Name, chall.Result)
		}
	}

	return nil
}
