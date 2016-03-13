package slack

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

const APIBaseURL = "https://slack.com/api"

// SlackTeam is a slack registered team
type SlackTeam struct {
	TeamID      string `json:"id"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	EmailDomain string `json:"email_domain"`
}

// Attachment is meant for extra text or image in slack response
type Attachment struct {
	Title    string `json:"title"`
	Text     string `json:"text"`
	Fallback string `json:"fallback"`
	ImageURL string `json:"image_url"`
	Color    string `json:"color"`
}

// ResponseMessage is slack response for the actions
type ResponseMessage struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type SlackTeamResponse struct {
	Team SlackTeam `json:"team"`
}

func GetTeamInfo(client *http.Client, token *oauth2.Token) (*SlackTeamResponse, error) {
	response, err := client.Get(fmt.Sprintf("%s/team.info?token=%s", APIBaseURL, token.AccessToken))

	if err != nil {
		log.Printf("Could not get the user information based on the token %s\n", err)
		return nil, err
	}
	defer response.Body.Close()

	var teamResponse SlackTeamResponse
	err = json.NewDecoder(response.Body).Decode(&teamResponse)
	if err != nil {
		return nil, err
	}

	return &teamResponse, nil
}
