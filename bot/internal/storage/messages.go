package storage

import (
	"database/sql"
	"log"
	"strings"
	"time"
)

type Message struct {
	Nick      string
	Content   string
	Channel   string
	Timestamp time.Time
}

func SubmitMessages(messages []Message, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	userInsertionStmt, err := tx.Prepare("INSERT OR IGNORE INTO users(nick, registered, opt, deletion) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}

	messagesStmt, err := tx.Prepare("INSERT INTO messages (nick, channel, message, time) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}

	for _, message := range messages {
		_, err = userInsertionStmt.Exec(message.Nick, message.Timestamp, false, nil)
		if err != nil {
			tx.Rollback()
			return err
		}

		_, err := messagesStmt.Exec(message.Nick, message.Channel, strings.TrimSpace(message.Content), message.Timestamp)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func FulfilsMessagesCount(nick string, quota int, db *sql.DB) (bool, int) {
	var count int
	var currentCount int

	err := db.QueryRow("SELECT COUNT(*) FROM messages WHERE nick = ?", nick).Scan(&count)
	if err != nil {
		log.Printf("Failed to count messages in FulfilsMessagesCount for nick %s: %s\n", nick, err.Error())
		return false, 0
	}

	err = db.QueryRow("SELECT COUNT(*) FROM messages WHERE nick = ?", nick).Scan(&currentCount)
	if err != nil {
		log.Printf("Failed to count individual messages in FulfilsMessagesCount: %s\n", err.Error())
		return false, 0
	}

	return (count >= quota), currentCount
}

func EnoughFulfilsMessagesCount(peopleQuota int, messageQuota int, db *sql.DB) bool {
	var count int

	err := db.QueryRow("SELECT COUNT(*) FROM (SELECT nick FROM messages GROUP BY nick HAVING COUNT(*) >= ?)", messageQuota).Scan(&count)
	if err != nil {
		log.Printf("Failed to count messages in EnoughFulfilsMessagesCount: %s\n", err.Error())
	}

	return (count >= peopleQuota)
}
