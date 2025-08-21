package commands

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"hearsay/internal/config"
	"hearsay/internal/storage"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/shlex"
)

func _boolToInt(a bool) int {
	if a {
		return 1
	}

	return 0
}

type retrainResponse struct {
	TimeD    float64 `json:"time"`
	Url      string  `json:"url"`
	Accuracy float64 `json:"accuracy"`
	F1       float64 `json:"f1"`
}

var lastRetrain = time.Now().Add(-2 * time.Hour)

func retrainHandler(args []string, author string, db *sql.DB) string {
	if !storage.IsOptedIn(author) {
		return fmt.Sprintf("%s: You must be opted in to use this command. %shelp opt", author, config.CommandPrefix)
	}

	if !storage.EnoughFulfilsMessagesCount(config.PeopleQuota, config.MessageQuota, db) {
		return fmt.Sprintf("%s: Not enough people fulfil the message quota. hearsay requires %d people with >= %d messages", author, config.PeopleQuota, config.MessageQuota)
	}

	if time.Since(lastRetrain) < 2*time.Hour {
		return author + ": The model has already been retrained within the last 2 hours"
	}
	lastRetrain = time.Now()

	url := fmt.Sprintf("http://api:8111/retrain?min_messages=%d", config.MessageQuota)
	if len(args) != 0 {
		inArgs, err := shlex.Split(strings.Join(args, " "))
		if err != nil {
			log.Printf("shlex failed to split arguments in retrain. (query: %s): %s", strings.Join(args, " "), err.Error())
			return fmt.Sprintf("%s: Failed to parse arguments (%s)", author, err.Error())
		} else {
			fs := flag.NewFlagSet("retrainArgs", flag.ContinueOnError)

			cm := fs.Bool("cm", false, "...")
			past := fs.Int("past", 0, "...")
			bert := fs.Bool("bert", false, "...")
			if *bert && !config.Bert {
				return fmt.Sprintf("%s: BERT has been disabled by the administrator", author)
			}

			err = fs.Parse(inArgs)
			if err != nil {
				log.Printf("shlex failed to parse arguments in retrain. (query: %s): %s", strings.Join(args, " "), err.Error())
				return fmt.Sprintf("%s: Failed to parse arguments (%s)", author, err.Error())
			} else {
				url = fmt.Sprintf("http://api:8111/retrain?cm=%d&cf=%d&bert=%d&min_messages=%d&gpu=%d", _boolToInt(*cm), *past, _boolToInt(*bert), config.MessageQuota, _boolToInt(config.GPU))
			}
		}
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Failed to get retrain URL for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results."
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to send GET request in retrain for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results."
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read response body in retrain for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results."
	}

	var result retrainResponse
	err = json.Unmarshal(resBody, &result)
	if err != nil {
		log.Printf("Failed to unmarshal response body in retrain for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results."
	}

	responseOne := fmt.Sprintf("%s: The SVM model has been retrained. It took \x02%.2f\x02 seconds to fit.", author, result.TimeD)
	if result.Url != "" {
		responseOne += fmt.Sprintf(" \x02Confusion matrix\x02: %s | \x025-fold CV\x02: Accuracy %.4f, F1 score %.4f", result.Url, result.Accuracy, result.F1)
	}

	return responseOne
}

var retrainHelp string = `Refit the classification model. This can be done every 2 hours. Add the --cm flag for evaluation statistics (heavy). To ignore inactive nicks, provide the --past flag with the number of days of inactivity before being cut off. To include BERT embeddings, append the --bert flag. NOTE: Using BERT is very slow with minimal accuracy gain. This is compounded when used in conjunction with --cm. Usage: ` + config.CommandPrefix + `retrain [--cm, --bert, --past <days>]`
