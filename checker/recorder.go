package checker

/***
* This file implements helper functions of mysql database.
***/

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

/***
* Data schema of test result table.
* Note that this is different from `Challenge` structure, which also contains test result.
***/
type DbResult struct {
	ChallId   int        `db:"challid"`
	Name      string     `db:"name"`
	Result    TestResult `db:"result"`
	Timestamp time.Time  `db:"timestamp"`
}

/***
* Converter from `Challenge` structure into `DBResult`.
***/
func (chall *Challenge) intoDbResult() DbResult {
	return DbResult{
		ChallId: chall.Id,
		Name:    chall.Name,
		Result:  chall.Result,
	}
}

/***
* Connect to mysql server and returns instance.
***/
func Connect(dbuser string, dbpass string, dbhost string, dbname string) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("%s:%s@(%s)/%s?parseTime=true&autocommit=0", dbuser, dbpass, dbhost, dbname)
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

/***
* Write and commit test result.
***/
func RecordResult(db *sqlx.DB, chall Challenge) error {
	tx := db.MustBegin()
	dbresult := chall.intoDbResult()
	dbresult.Timestamp = time.Now()
	query := "insert into test_result(challid, name, result, timestamp) values(:challid, :name, :result, :timestamp)"
	_, err := tx.NamedExec(query, dbresult)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

/***
* Query test result from DB by challenge ID.
***/
func FetchResult(db *sqlx.DB, challid int, limit int) ([]DbResult, error) {
	var results []DbResult

	query := `select challid, name, result, timestamp from test_result where challid = ? order by timestamp desc limit ?`
	tx := db.MustBegin()
	if err := tx.Select(&results, query, challid, limit); err != nil {
		return results, err
	}
	if err := tx.Commit(); err != nil {
		return results, err
	}
	return results, nil
}
