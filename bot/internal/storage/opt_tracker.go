package storage

import (
	"database/sql"
)

// We keep a map of opt-outs. This prevents database lookups.
// This approach works fine for smaller servers.

var OptIns = make(map[string]struct{}) // TODO: Add mutex if this is used in more than one place.

func IsOptedIn(nick string) bool {
	if _, exists := OptIns[nick]; exists {
		return true
	}
	return false
}

func LoadOptIns(db *sql.DB) error {
	res, err := db.Query("SELECT nick FROM users WHERE opt = 1")
	if err != nil {
		return err
	}
	defer res.Close()

	for res.Next() {
		var nick string
		if err := res.Scan(&nick); err != nil {
			return err
		}
		OptIns[nick] = struct{}{}
	}

	return nil
}
