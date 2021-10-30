package badge

import (
	"fmt"
	"testing"

	"github.com/smallkirby/skbctf-status/checker"
)

type Challenge = checker.Challenge

func TestBadgeUrl(t *testing.T) {
	dbuser := "testuser"
	dbpass := "testpass"
	dbhost := "localhost"
	dbname := "testname"

	// first, record a challenge
	chall := Challenge{
		Name:            "TestBadgeUrl Challenge",
		Id:              0,
		Default_success: true,
		Result:          checker.TestSuccess,
	}
	db, err := checker.Connect(dbuser, dbpass, dbhost, dbname)
	if err != nil {
		t.Errorf("Failed to connect to DB:\n%v", err)
	}
	// record to DB
	if err := checker.RecordResult(db, chall); err != nil {
		t.Errorf("%v", err)
	}

	// fetch badge URL
	badger, err := NewBadger(dbuser, dbpass, dbhost, dbname)
	if err != nil {
		t.Errorf("%v", err)
	}

	url, err := badger.GetBadge(0)
	if err != nil {
		t.Errorf("%v", err)
	}

	fmt.Printf("Shields URL: %s\n", url)
}
