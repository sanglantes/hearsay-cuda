package commands

import "database/sql"

type CommandFunc func(args []string, author string, db *sql.DB) string
type Command struct {
	Handler     CommandFunc
	Description string
}

var Commands = make(map[string]Command)

func init() {
	Commands["attribute"] = Command{attributeHandler, attributeHelp}
	Commands["opt"] = Command{optHandler, optHelp}
	Commands["forget"] = Command{forgetHandler, forgetHelp}
	Commands["unforget"] = Command{unforgetHandler, unforgetHelp}
	Commands["help"] = Command{helpHandler, helpHelp}
	Commands["readability"] = Command{readabilityHandler, readabilityHelp}
	Commands["retrain"] = Command{retrainHandler, retrainHelp}
	Commands["about"] = Command{aboutHandler, aboutHelp}
	Commands["me"] = Command{meHandler, meHelp}
	Commands["sentiment"] = Command{sentimentHandler, sentimentHelp}
	Commands["profile"] = Command{profileHandler, profileHelp}
}
