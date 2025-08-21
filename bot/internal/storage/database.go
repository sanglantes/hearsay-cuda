package storage

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func openDbHelper(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatalf("Error opening %s: %v\n", path, err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to %s: %v\n", path, err)
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatalf("Failed to enable foreign keys: %s\n", err.Error())
	}

	return db
}

func InitDatabase() (*sql.DB, error) {
	db := openDbHelper("data/database.db")
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS messages(
	id INTEGER PRIMARY KEY,
	nick TEXT NOT NULL,
	channel TEXT NOT NULL,
	message TEXT NOT NULL,
	time DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY(nick) REFERENCES users(nick) ON DELETE CASCADE
	)`)
	if err != nil {
		log.Fatalf("Error creating messages table: %v\n", err.Error())
		return nil, err
	}

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_nick ON messages(nick)")
	if err != nil {
		log.Fatalf("Error indexing nick: %v\n", err.Error())
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users(
	nick TEXT PRIMARY KEY,
	registered DATETIME DEFAULT CURRENT_TIMESTAMP,
	opt BOOL DEFAULT FALSE,
	deletion DATETIME
	)`)
	if err != nil {
		log.Fatalf("Error creating users table: %v\n", err.Error())
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS profiles(
	id INTEGER PRIMARY KEY,
	nick TEXT,
	name TEXT,
	messages TEXT DEFAULT "",
	created DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY(nick) REFERENCES users(nick) ON DELETE CASCADE
	)`)
	if err != nil {
		log.Fatalf("Error creating profiles table: %v\n", err.Error())
		return nil, err
	}

	return db, nil
}
