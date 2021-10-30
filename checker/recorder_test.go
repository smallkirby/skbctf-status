package checker

import (
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestRecordFetchResult(t *testing.T) {
	dbuser := "testuser"
	dbpass := "testpass"
	dbhost := "localhost"
	dbname := "testname"
	chall := Challenge{
		Name:            "Test Chall",
		Id:              0,
		Default_success: false,
		Result:          TestSuccess,
	}

	db, err := Connect(dbuser, dbpass, dbhost, dbname)
	if err != nil {
		t.Errorf("Failed to connect to DB:\n%v", err)
	}
	// record to DB
	if err := RecordResult(db, chall); err != nil {
		t.Errorf("%v", err)
	}
	// fetch from DB
	results, err := FetchResult(db, chall.Id, 3)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(results) == 0 {
		t.Error("Fetched record size is 0.")
	} else {
		fmt.Printf("Fetched record size is %d.\n", len(results))
	}
}
