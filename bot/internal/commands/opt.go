package commands

import (
	"database/sql"
	"fmt"
	"hearsay/internal/config"
	"hearsay/internal/storage"
	"log"
)

func optHandler(args []string, author string, db *sql.DB) string {
	opt := map[string]bool{
		"in":  true,
		"out": false,
	}

	if len(args) == 0 {
		optReverse := map[bool]string{
			true:  "in",
			false: "out",
		}
		var optBool bool
		err := db.QueryRow("SELECT opt FROM users WHERE nick = ?", author).Scan(&optBool)

		if err == sql.ErrNoRows {
			return author + ": Your nick was not found in the database"
		} else if err != nil {
			log.Printf("%s", err.Error())
			return author + ": Something went wrong"
		}

		return fmt.Sprintf("%s: You are currently opted %s.", author, optReverse[optBool])
	}

	if len(args) != 1 || (args[0] != "in" && args[0] != "out") {
		return author + ": Improper argument(s). See " + config.CommandPrefix + "help opt for usage."
	}

	res, err := db.Exec("UPDATE users SET opt = ? WHERE nick = ?", opt[args[0]], author)
	if err != nil {
		log.Printf("Failed updating opt preference: %s\n", err.Error())
		return author + ": Something went wrong"
	}

	rA, _ := res.RowsAffected()
	if rA == 0 {
		log.Println("In opt command: user not found")
		return author + ": Your nick was not found in the database"
	}

	if opt[args[0]] {
		delete(storage.OptIns, author)
	} else {
		storage.OptIns[author] = struct{}{}
	}
	return author + ": You have successfully opted " + args[0] + "."
}

var optHelp string = `Opt in or out from data collection and model training. If no arguments are submitted, your current opt status will be returned. Usage: ` + config.CommandPrefix + `opt [in|out]. (default: out)`
