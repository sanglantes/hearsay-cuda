package commands

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"hearsay/internal/config"
	"hearsay/internal/storage"
	"io"
	"log"
	"net/http"
	"strings"
)

type sentimentResponse struct {
	Pos           float64 `json:"pos"`
	Neu           float64 `json:"neu"`
	Neg           float64 `json:"neg"`
	HumanReadable string  `json:"hr"`
	Compound      float64 `json:"compound"`
}

func sentimentHandler(args []string, author string, db *sql.DB) string {
	if !storage.IsOptedIn(author) {
		return fmt.Sprintf("%s: You must be opted in to use this command. %shelp opt", author, config.CommandPrefix)
	}

	if len(args) == 0 {
		return author + ": You cannot submit an empty message"
	}

	msg := strings.Join(args, " ")
	body := map[string]interface{}{
		"msg": msg,
	}
	postJson, err := json.Marshal(body)
	if err != nil {
		log.Printf("Failed to marshal POST request in sentiment by %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}

	req, err := http.NewRequest(http.MethodPost, "http://api:8111/sentiment", bytes.NewBuffer(postJson))
	if err != nil {
		log.Printf("Failed to get sentiment URL for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to send POST request in sentiment for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results."
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read response body in sentiment for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}

	var result sentimentResponse
	err = json.Unmarshal(resBody, &result)
	if err != nil {
		log.Printf("Failed to unmarshal response body in sentiment for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}

	return fmt.Sprintf("%s: Largely \x02%s\x02 with a compound score of \x02%.2f\x02. (pos: %.2f, neu: %.2f, neg: %.2f)", author, result.HumanReadable, result.Compound, result.Pos, result.Neu, result.Neg)
}

var sentimentHelp string = `Extract the sentiment (positive, neutral, or negative) from a message. Usage: ` + config.CommandPrefix + `sentiment <message>`
