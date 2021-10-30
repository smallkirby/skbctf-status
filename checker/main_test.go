package checker

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

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
		if f.IsDir() && f.Name() != "chall3" {
			res := make(chan Challenge, 1)
			abspath, _ := filepath.Abs(filepath.Join("../examples", f.Name()))
			executer := Executer{path: abspath, logger: *slogger}
			executer.Check(res, "info.json")
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
		Timeout:   3.0,
		Infofile:  "info.json",
		Nodb:      true,
		ChallsDir: "../examples",
	}

	start_time := time.Now().Unix()
	if err := CheckAllOnce(*slogger, conf); err != nil {
		t.Errorf("Test failed: %v", err)
	}
	end_time := time.Now().Unix()
	fmt.Printf("Total time: %v.", end_time-start_time)
}

func TestAsyncAllCheck(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	slogger := logger.Sugar()
	conf := CheckerConfig{
		Single:    true,
		Parallel:  true,
		Timeout:   3.0,
		Infofile:  "info.json",
		Nodb:      true,
		ChallsDir: "../examples",
	}

	start_time := time.Now().Unix()
	if err := CheckAllOnce(*slogger, conf); err != nil {
		t.Errorf("Test failed: %v", err)
	}
	end_time := time.Now().Unix()
	fmt.Printf("Total time: %v.", end_time-start_time)
}

func TestRecordCheckAllOnce(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	slogger := logger.Sugar()
	conf := CheckerConfig{
		Single:    true,
		Parallel:  false,
		Timeout:   3.0,
		Infofile:  "info.json",
		Nodb:      false,
		ChallsDir: "../examples",
	}

	dbhost := "localhost"
	dbname := "testname"
	dbuser := "testuser"
	dbpass := "testpass"
	os.Setenv("DBNAME", dbname)
	os.Setenv("DBUSER", dbuser)
	os.Setenv("DBPASS", dbpass)
	os.Setenv("DBHOST", dbhost)

	start_time := time.Now().Unix()
	if err := CheckAllOnce(*slogger, conf); err != nil {
		t.Errorf("Test failed: %v", err)
	}
	end_time := time.Now().Unix()
	fmt.Printf("Total time: %v.\n", end_time-start_time)

	db, err := Connect(os.Getenv("DBUSER"), os.Getenv("DBPASS"), os.Getenv("DBHOST"), os.Getenv("DBNAME"))
	if err != nil {
		t.Errorf("Failed to connect to DB: \n%v", err)
	}

	result, err := FetchResult(db, 0, 1)
	if err != nil {
		t.Errorf("Failed to fetch records: \n%v", err)
	}
	if result[0].ChallId != 0 || result[0].Result != TestSuccessWithoutExecution {
		t.Errorf("Invalid record returned: %v", result[0])
	}

	result, err = FetchResult(db, 1, 1)
	if err != nil {
		t.Errorf("Failed to fetch records: \n%v", err)
	}
	if result[0].ChallId != 1 || result[0].Result != TestSuccess {
		t.Errorf("Invalid record returned: %v", result[0])
	}

	result, err = FetchResult(db, 2, 1)
	if err != nil {
		t.Errorf("Failed to fetch records: \n%v", err)
	}
	if result[0].ChallId != 2 || result[0].Result != TestFailure {
		t.Errorf("Invalid record returned: %v", result[0])
	}

	result, err = FetchResult(db, 3, 1)
	if err != nil {
		t.Errorf("Failed to fetch records: \n%v", err)
	}
	if result[0].ChallId != 3 || result[0].Result != TestTimeout {
		t.Errorf("Invalid record returned: %v", result[0])
	}

	result, err = FetchResult(db, 999, 1)
	if err != nil {
		t.Errorf("Failed to fetch records: \n%v", err)
	}
	if len(result) != 0 {
		t.Errorf("Result which mustn't exist is returned: %v", result[0])
	}
}
