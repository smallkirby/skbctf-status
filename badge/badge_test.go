package badge

import (
	"fmt"
	"testing"
)

func TestBadgeUrl(t *testing.T) {
	dbuser := "testuser"
	dbpass := "testpass"
	dbhost := "localhost"
	dbname := "testname"
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
