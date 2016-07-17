package controller

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/slack-games/slack-client"
)

// CommandInput user input for the game commands
type CommandInput struct {
	ChannelName string `schema:"channel_name" validate:"required"`
	ChannelID   string `schema:"channel_id" validate:"required,alphanum"`
	TeamID      string `schema:"team_id" validate:"required,alphanum"`
	UserID      string `schema:"user_id" validate:"required,alphanum"`
	Text        string `schema:"text" validate:"required"`
	Domain      string `schema:"team_domain" validate:"required"`
	Name        string `schema:"user_name" validate:"required"`
}

func sendResponse(w http.ResponseWriter, message slack.ResponseMessage) {
	// Set headers
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Generate message output
	if err := json.NewEncoder(w).Encode(message); err != nil {
		log.Fatal("Could not generate the message json", err)
	}
}

func slackTokenHandler(token string) func(next http.Handler) http.Handler {
	// Get function handler
	return func(next http.Handler) http.Handler {
		// Response - request
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			slackToken := r.PostFormValue("token")

			if token != slackToken {
				sendResponse(w, slack.ResponseMessage{
					Text:        "Make sure the Slack APP tokens are same",
					Attachments: []slack.Attachment{},
				})
				return
			}
			log.Println("Valid slack token")
			next.ServeHTTP(w, r)
		})
	}
}

func debugFormValues(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Show keys for debugging
		log.Println("> Form debug starts here")
		for key, value := range r.PostForm {
			log.Printf("\tkey: %s value: %s\n", key, value)
		}
		log.Println("> Headers ends here")
		next.ServeHTTP(w, r)
	})
}

func randomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
