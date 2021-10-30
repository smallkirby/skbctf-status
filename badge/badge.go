package badge

import (
	"fmt"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/smallkirby/skbctf-status/checker"
)

type Badger struct {
	db *sqlx.DB
}

func NewBadger() (*Badger, error) {
	if db, err := checker.Connect(os.Getenv("DBUSER"), os.Getenv("DBPASS"), os.Getenv("DBHOST"), os.Getenv("DBNAME")); err != nil {
		return nil, err
	} else {
		return &Badger{db: db}, nil
	}
}

func toShieldsString(s string) string {
	/*

		Dashes --	→	- Dash
		Underscores __	→	_ Underscore
		_ or Space  	→	  Space

	*/
	s = strings.ReplaceAll(s, "-", "--")
	s = strings.ReplaceAll(s, "_", "__")
	return s
}

func toShieldsUrl(label string, message string, color string) string {
	label = toShieldsString(label)
	message = toShieldsString(message)
	value := fmt.Sprintf("%s-%s-%s", label, message, color)
	return fmt.Sprintf("https://img.shields.io/badge/%s", value)
}

// https://img.shields.io/badge/<LABEL>-<MESSAGE>-<COLOR>
func (bd Badger) GetBadge(challid int) (string, error) {
	results, err := checker.FetchResult(bd.db, challid, 1)
	if err != nil {
		return "", err
	}

	if len(results) != 1 {
		return "", fmt.Errorf("Status for %v not found.", challid)
	}
	result := results[0]

	status := result.Result
	message := status.ToMessage()
	label := result.Name
	color := status.ToColor()
	url := toShieldsUrl(label, message, color)

	return url, nil
}
