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

type scoreResponse struct {
	Score float64 `json:"score"`
}

func scoreClass(score float64) string {
	switch {
	case 90.0 <= score:
		return "5th grade level. Very easy to read."
	case 80.0 <= score:
		return "6th grade level. Easy to read. Conversational English for consumers."
	case 70.0 <= score:
		return "7th grade level. Fairly easy to read."
	case 60.0 <= score:
		return "8th & 9th grade. Plain English."
	case 50.0 <= score:
		return "10th to 12th grade. Fairly difficult to read."
	case 30.0 <= score:
		return "College level. Difficult to read."
	case 10.0 <= score:
		return "College graduate level. Very difficult to read."
	case 0.0 <= score:
		return "Professional level. Extremely difficult to read."
	}

	return "Unknown."
}

func readabilityHandler(args []string, author string, db *sql.DB) string {
	if !storage.IsOptedIn(author) {
		return fmt.Sprintf("%s: You must be opted in to use this command. %shelp opt", author, config.CommandPrefix)
	}

	fulfil, count := storage.FulfilsMessagesCount(author, config.MessageQuota, db)
	if !fulfil {
		return fmt.Sprintf("%s: You have too few messages stored to use this command (%d/%d required)", author, count, config.MessageQuota)
	}

	url := fmt.Sprintf("http://api:8111/readability?nick=%s", author)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Failed to get readability URL for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to send GET request in readability for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read response body in readability for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}
	var result scoreResponse
	err = json.Unmarshal(resBody, &result)
	if err != nil {
		log.Printf("Failed to unmarshal response body in readability for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}

	return fmt.Sprintf("%s: You have a Flesch-Kincaid score of %.2f (%s)", author, result.Score, scoreClass(result.Score))
}

var readabilityHelp string = `Calculate the Flesch-Kincaid readability score of your messages (10,000 limit). Usage: ` + config.CommandPrefix + `readability`
