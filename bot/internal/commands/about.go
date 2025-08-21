package commands

import (
	"database/sql"
	"hearsay/internal/config"
)

func aboutHandler(args []string, author string, db *sql.DB) string {
	return author + ": hearsay is an authorship attribution and natural language processing bot made in Go and Python. It works by training a machine learning model on stylometric features such as word length frequency, punctuation, character n-grams, and capitalization. The model can then predict likely authors from unknown messages and compare similarities between authors. You must manually opt in to use hearsay. Get help with " + config.CommandPrefix + "help."
}

var aboutHelp string = `Information about hearsay. Usage: ` + config.CommandPrefix + `about`
