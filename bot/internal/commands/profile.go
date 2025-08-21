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

var maxProfiles int = 3

func exceedsMaxProfiles(author string, db *sql.DB) bool {
	var res int
	err := db.QueryRow("SELECT COUNT(*) FROM profiles WHERE nick = ?", author).Scan(&res)
	if err != nil {
		return true
	}

	return (res >= maxProfiles)
}

func profileExists(name string, author string, db *sql.DB) (bool, error) {
	var res int
	err := db.QueryRow("SELECT COUNT(*) FROM profiles WHERE nick = ? AND name = ?", author, name).Scan(&res)
	if err != nil {
		return true, err
	}

	return (res > 0), nil
}

func listProfile(args []string, author string, db *sql.DB) string {
	res, err := db.Query("SELECT name, messages FROM profiles WHERE nick = ?", author)
	if err != nil {
		log.Printf("Failed to query profiles by %s: %s", author, err.Error())
		return fmt.Sprintf("%s: Failed to fetch results", author)
	}
	defer res.Close()

	profileNames := make(map[string]int)
	for res.Next() {
		var name string
		var msg string
		if err := res.Scan(&name, &msg); err != nil {
			log.Printf("Failed to scan profiles and messages by %s: %s", author, err.Error())
			return fmt.Sprintf("%s: Failed to fetch results", author)
		}

		count := len(strings.Split(msg, "/:MSG/"))
		profileNames[name] = max(0, count-1)
	}

	result := fmt.Sprintf("%s: You have %d profiles:", author, len(profileNames))
	for k, v := range profileNames {
		result += fmt.Sprintf(" %s (%d messages),", k, v)
	}

	result = strings.TrimSuffix(result, ",")

	return result
}

func createProfile(args []string, author string, db *sql.DB) string {
	if len(args) != 2 {
		return fmt.Sprintf("%s: Too few or too many arguments supplied. Profile names may not contain spaces", author)
	}

	if exceedsMaxProfiles(author, db) {
		return fmt.Sprintf("%s: You have reached the maximum number of profiles allowed (%d). Delete a profile before continuing", author, maxProfiles)
	}

	exists, _ := profileExists(args[1], author, db)
	if exists {
		return fmt.Sprintf("%s: A profile called %s already exists in your name", author, args[1])
	}

	_, err := db.Exec("INSERT INTO profiles (nick, name, messages) VALUES (?, ?, \"\")", author, args[1])
	if err != nil {
		log.Printf("Failed to create profile for %s: %s given %v", author, err.Error(), args)
		return fmt.Sprintf("%s: Failed to create profile", author)
	}

	return fmt.Sprintf("%s: You have created a new profile %s", author, args[1])
}

func destroyProfile(args []string, author string, db *sql.DB) string {
	if len(args) != 2 {
		return fmt.Sprintf("%s: Too few or too many arguments supplied. Profile names may not contain spaces", author)
	}

	exists, _ := profileExists(args[1], author, db)
	if !exists {
		return fmt.Sprintf("%s: No profile called %s exists in your name", author, args[1])
	}

	_, err := db.Exec("DELETE FROM profiles WHERE nick = ? AND name = ?", author, args[1])
	if err != nil {
		log.Printf("Failed to delete profile for %s: %s given %v", author, err.Error(), args)
		return fmt.Sprintf("%s: Failed to delete profile", author)
	}

	return fmt.Sprintf("%s: You have deleted the profile %s", author, args[1])
}

func appendProfile(args []string, author string, db *sql.DB) string {
	if len(args) < 3 {
		return fmt.Sprintf("%s: Too few arguments supplied", author)
	}

	exists, _ := profileExists(args[1], author, db)
	if !exists {
		return fmt.Sprintf("%s: No profile called %s exists in your name", author, args[1])
	}

	_, err := db.Exec("UPDATE profiles SET messages = messages || '/:MSG/' || ? WHERE nick = ? AND name = ?", strings.Join(args[2:], " "), author, args[1])
	if err != nil {
		log.Printf("Failed to append message to profile for %s: %s given %v", author, err.Error(), args)
		return fmt.Sprintf("%s: Failed to append message to profile", author)
	}

	return ""
}

func getMessagesFromProfile(name string, author string, db *sql.DB) (string, error) {
	var messages string
	err := db.QueryRow("SELECT messages FROM profiles WHERE nick = ? AND name = ?", author, name).Scan(&messages)
	if err != nil {
		log.Printf("Failed to query profiles by %s: %s", author, err.Error())
		return "", err
	}

	return messages, nil
}

func attributeProfile(args []string, author string, db *sql.DB) string {
	if len(args) != 2 {
		return fmt.Sprintf("%s: Too few or too many arguments supplied. Profile names may not contain spaces", author)
	}

	msg, err := getMessagesFromProfile(args[1], author, db)
	body := map[string]interface{}{
		"msg":          msg,
		"min_messages": config.MessageQuota,
		"confidence":   true,
	}
	postJson, err := json.Marshal(body)
	if err != nil {
		log.Printf("Failed to marshal POST request in profile attribute by %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}

	req, err := http.NewRequest(http.MethodPost, "http://api:8111/profile_attribute", bytes.NewBuffer(postJson))
	if err != nil {
		log.Printf("Failed to get profile attribute URL for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to send POST request in profile attribute for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read response body in profile attribute for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}

	var result attributeResponse
	err = json.Unmarshal(resBody, &result)
	if err != nil {
		log.Printf("Failed to unmarshal response body in profile attribute for %s: %s\n", author, err.Error())
		return author + ": Failed to fetch results"
	}

	return fmt.Sprintf("%s: Predicted author: %s_. Confidence scores: %s", author, result.Author, result.ConfidenceScore)
}

func profileHandler(args []string, author string, db *sql.DB) string {
	if !storage.IsOptedIn(author) {
		return fmt.Sprintf("%s: You must be opted in to use this command. %shelp opt", author, config.CommandPrefix)
	}

	if len(args) < 1 {
		return fmt.Sprintf("%s: No arguments supplied. See %shelp profile", author, config.CommandPrefix)
	}

	type fn func([]string, string, *sql.DB) string
	argumentFuncMap := map[string]fn{
		"list":      listProfile,
		"create":    createProfile,
		"destroy":   destroyProfile,
		"append":    appendProfile,
		"attribute": attributeProfile,
	}
	if profileFunction, ok := argumentFuncMap[args[0]]; ok {
		return profileFunction(args, author, db)
	}

	return fmt.Sprintf("%s: Invalid argument: %s. See %shelp profile.", author, args[0], config.CommandPrefix)
}

var profileHelp string = `Build author profiles that provide higher attribution accuracy. Usage: ` + config.CommandPrefix + `profile (attribute|create|destroy) <name> | append <name> <message> | list`
