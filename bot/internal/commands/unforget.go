package commands

import (
	"database/sql"
	"hearsay/internal/config"
	"log"
)

func unforgetHandler(args []string, author string, db *sql.DB) string {
	res, err := db.Exec("UPDATE users SET deletion = NULL WHERE nick = ? AND deletion IS NOT NULL", author)
	if err != nil {
		log.Printf("Failed to serve unforget request for nick %s: %s\n", author, err.Error())
		return author + ": The requested action was met with an error."
	}

	if rA, _ := res.RowsAffected(); rA == 0 {
		return author + ": You have no deletion scheduled or were not found in the database."
	}

	return author + ": You have successfully cancelled your deletion request."
}

var unforgetHelp string = `Cancel a scheduled data deletion. Usage: ` + config.CommandPrefix + `unforget`
