package commands

import (
	"context"
	"database/sql"
	"hearsay/internal/config"
	"log"
	"strconv"
	"time"

	irc "github.com/fluffle/goirc/client"
)

func forgetHandler(args []string, author string, db *sql.DB) string {
	var deletionTime sql.NullTime
	res := db.QueryRow("SELECT deletion FROM users WHERE nick = ?", author)
	if err := res.Scan(&deletionTime); err != nil {
		if err == sql.ErrNoRows {
			return author + ": Your nick was not found in the database"
		}
		log.Printf("Failed to query deletion time: %s\n", err.Error())
		return author + ": The requested action was met with an error"
	}

	if deletionTime.Valid {
		return author + ": Your data is already scheduled for deletion"
	}

	deletionDate := time.Now()
	deletionDate = deletionDate.AddDate(0, 0, config.DeletionDays)
	deletionDate = deletionDate.Truncate(24 * time.Hour) // Truncating anything past the day.
	_, err := db.Exec("UPDATE users SET deletion = ? WHERE nick = ?", deletionDate, author)
	if err != nil {
		log.Printf("Failed to schedule deletion: %s\n", err.Error())
		return author + ": The requested action was met with an error"
	}

	return author + ": Your data is scheduled for deletion and will complete in " + strconv.Itoa(config.DeletionDays) + " days. To cancel this request, type +unforget"
}

var forgetHelp string = `Permanently purge all your data. Usage: ` + config.CommandPrefix + `forget`

func deletionExecuter(db *sql.DB) []string {
	// TODO: Transform to transaction.
	deletedNicks := []string{}
	res, err := db.Query("SELECT nick FROM users WHERE DATE(deletion) = DATE('now')")
	if err != nil {
		log.Printf("Failed to query today's deletions: %s\n", err.Error())
		return make([]string, 0)
	}
	defer res.Close()

	for res.Next() {
		var nick string
		err := res.Scan(&nick)
		if err != nil {
			log.Printf("Failed to scan to-be-deleted nick: %s\n", err.Error())
		}

		_, err = db.Exec("DELETE FROM users WHERE nick = ?", nick)
		if err != nil {
			log.Printf("Failed to delete nick from users table: %s\n", err.Error())
		} else {
			deletedNicks = append(deletedNicks, nick)
		}
	}

	return deletedNicks
}

func DeletionWrapper(db *sql.DB, c *irc.Conn, ctx context.Context) {
	for {
		now := time.Now()
		next := now.Truncate(24 * time.Hour).Add(24 * time.Hour)
		log.Printf("Next deletion cycle scheduled for %v\n", next)

		select {
		case <-ctx.Done():
			log.Println("Shutting down scheduler.")
			return

		case <-time.After(time.Until(next)):
			deletedNicks := deletionExecuter(db)
			for _, nick := range deletedNicks {
				// TODO: If the user isn't online, postpone the reminder until they are.
				c.Privmsg(nick, "Your data has been successfully purged")
			}
		}
	}
}
