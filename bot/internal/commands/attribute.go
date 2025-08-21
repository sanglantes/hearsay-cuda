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

type attributeResponse struct {
	Author          string `json:"author"`
	ConfidenceScore string `json:"confidence"`
}

type attributeListResponse struct {
	Authors string `json:"authors"`
}

func attributeHandler(args []string, author string, db *sql.DB) string {
	if !storage.IsOptedIn(author) {
		return fmt.Sprintf("%s: You must be opted in to use this command. %shelp opt", author, config.CommandPrefix)
	}

	if !storage.EnoughFulfilsMessagesCount(config.PeopleQuota, config.MessageQuota, db) {
		return fmt.Sprintf("%s: Not enough people fulfil the message quota. hearsay requires %d people with >= %d messages", author, config.PeopleQuota, config.MessageQuota)
	}

	if len(args) == 0 {
		return author + ": You cannot attribute an empty message"
	}

	if args[0] == "--list" {
		req, err := http.NewRequest(http.MethodGet, "http://api:8111/attribute_list", nil)
		if err != nil {
			log.Printf("Failed to get retrain URL for %s: %s\n", author, err.Error())
			return author + ": Failed to fetch results"
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Failed to send GET request in attribue for %s: %s\n", author, err.Error())
			return author + ": Failed to fetch results"
		}

		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			log.Printf("Failed to read response body in attribute for %s: %s\n", author, err.Error())
			return author + ": Failed to fetch results"
		}

		var result attributeListResponse
		err = json.Unmarshal(resBody, &result)
		if err != nil {
			log.Printf("Failed to unmarshal response body in attribute for %s: %s\n", author, err.Error())
			return author + ": Failed to fetch results"
		}

		return fmt.Sprintf("%s: Here is a list of nicks currently in the model's scope of view: %s", author, result.Authors)
	}

	msg := strings.Join(args, " ")
	body := map[string]interface{}{
		"msg":          msg,
		"min_messages": config.MessageQuota,
		"confidence":   true,
	}
	postJson, err := json.Marshal(body)
	if err != nil {
		log.Printf("Failed to marshal POST request in attribute by %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}

	req, err := http.NewRequest(http.MethodPost, "http://api:8111/attribute", bytes.NewBuffer(postJson))
	if err != nil {
		log.Printf("Failed to get attribute URL for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to send POST request in attribute for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read response body in attribute for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}

	var result attributeResponse
	err = json.Unmarshal(resBody, &result)
	if err != nil {
		log.Printf("Failed to unmarshal response body in attribute for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results."
	}

	return fmt.Sprintf("%s: Predicted author: %s_. Confidence scores: %s", author, result.Author, result.ConfidenceScore)
}

var attributeHelp string = `Attribute a message to a chatter who is opted in and fulfils the message quota. To view the model's scope of view, use the --list flag. NOTE: Longer messages will yield higher accuracy; aim for >= 10 characters. Usage: ` + config.CommandPrefix + `attribute (--list|<message>)`
