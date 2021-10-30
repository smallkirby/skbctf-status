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
		challs_wait_que := make([]Executer, len(challs))
		challs_running_num := 0
		ch := make(chan Challenge, len(challs))

		// init executer and push them into waiting que
		for ix, challdir := range challs {
			executer := Executer{path: challdir, logger: logger}
			challs_wait_que[ix] = executer
		}

		// exec all
		// XXX --asnyc-num option would specify # of running tests in a time
		for len(challs_wait_que) > 0 {
			chall := challs_wait_que[0]
			challs_wait_que = challs_wait_que[1:]
			go chall.CheckWithTimeout(ch, conf.Infofile, conf.Timeout)
			challs_running_num += 1
		}

		// wait and get results
		for result := range ch {
			logger.Infof("[%s] Test finish.", result.Name)
			challs_running_num -= 1
			// record to DB
			if !conf.Nodb {
				if err := RecordResult(db, result); err != nil {
					logger.Warn("%v", err)
				}
			}
			// end of tests
			if challs_running_num <= 0 {
				close(ch)
			}
		}
	} else {
		// Sequential test execution
		for _, challdir := range challs {
			ch := make(chan Challenge, 1)
			executer := Executer{path: challdir, logger: logger}
			go executer.CheckWithTimeout(ch, conf.Infofile, conf.Timeout)

			chall := <-ch
			if !conf.Nodb {
				if err := RecordResult(db, chall); err != nil {
					logger.Warn("%v", err)
				}
			}
			logger.Infof("[%s] Test finish: %v", chall.Name, chall.result)
		}
	}

	return nil
}
