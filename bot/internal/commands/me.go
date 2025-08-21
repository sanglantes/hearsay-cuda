package commands

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"hearsay/internal/config"
	"hearsay/internal/storage"
	"io"
	"log"
	"net/http"
)

type meResponse struct {
	ReadabilityScore float64 `json:"readability"`
	SentimentScore   float64 `json:"sentiment"`
	SentimentHR      string  `json:"sentiment_hr"`
	Neighbour        string  `json:"neighbour"`
}

func meHandler(args []string, author string, db *sql.DB) string {
	if !storage.IsOptedIn(author) {
		return fmt.Sprintf("%s: You must be opted in to use this command. %shelp opt", author, config.CommandPrefix)
	}

	fulfil, count := storage.FulfilsMessagesCount(author, config.MessageQuota, db)
	if !fulfil {
		return fmt.Sprintf("%s: You have too few messages stored to use this command (%d/%d required)", author, count, config.MessageQuota)
	}

	url := fmt.Sprintf("http://api:8111/me?author=%s", author)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Failed to get URL in me for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to send GET request in me for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read response body in me for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}

	var result meResponse
	err = json.Unmarshal(resBody, &result)
	if err != nil {
		log.Printf("Failed to unmarshal response body in retrain for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}

	return fmt.Sprintf("%s: Message count: \x02%d/%d\x02 | Readability: \x02%.2f\x02 | Sentiment: \x02%.2f\x02 (%s) | Neighbour: \x02%s\x02",
		author, count, config.MessageQuota, result.ReadabilityScore, result.SentimentScore, result.SentimentHR, result.Neighbour)
}

var meHelp string = `Statistics about yourself. Usage: ` + config.CommandPrefix + `me`
