package data

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"
)

func ImportLogs(db *sql.DB, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	checkStmt, err := tx.Prepare("SELECT 1 FROM messages WHERE nick = ? AND message = ? LIMIT 1")
	if err != nil {
		return err
	}
	userStmt, err := tx.Prepare("SELECT 1 FROM users WHERE nick = ? AND opt = 1 LIMIT 1")
	if err != nil {
		return err
	}
	messageStmt, err := tx.Prepare("INSERT INTO messages (nick, channel, message, time) VALUES (?, \"#antisocial\", ?, ?)")
	if err != nil {
		return err
	}
	defer messageStmt.Close()
	defer userStmt.Close()
	defer checkStmt.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.SplitN(scanner.Text(), " ", 5)
		if len(line) < 5 {
			continue
		}

		t, err := time.Parse("Jan 2 15:04:05 2006", fmt.Sprintf("%s %s %s %s", line[0], line[1], line[2], "2025"))
		if err != nil {
			return err
		}

		nick := line[3]
		message := line[4]
		var placeholder int

		err = checkStmt.QueryRow(nick, message).Scan(&placeholder)
		if err == nil {
			continue
		} else if err == sql.ErrNoRows {
			err = userStmt.QueryRow(nick).Scan(&placeholder)
			if err == sql.ErrNoRows {
				continue
			}
			_, err = messageStmt.Exec(nick, message, t)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return tx.Commit()
}
