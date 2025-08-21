package commands

import (
	"database/sql"
	"fmt"
	"hearsay/internal/config"
	"strings"
)

func helpHandler(args []string, author string, db *sql.DB) string {
	if len(args) == 0 {
		var listOfCommands []string
		for v := range Commands {
			listOfCommands = append(listOfCommands, v)
		}
		helpString := fmt.Sprintf(": Available commands are %s. Usage: %shelp [command]", strings.Join(listOfCommands, ", "), config.CommandPrefix)
		return author + helpString
	}

	key := args[0]

	if cmd, ok := Commands[key]; ok {
		return author + ": " + cmd.Description
	}

	return author + ": No such command " + args[0] + "."
}

var helpHelp string = `Get information on a command. Usage: ` + config.CommandPrefix + `help [command]`
