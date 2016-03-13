package commands

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/riston/slack-server/datastore"
	"github.com/riston/slack-server/slack"
)

// CurrentCommand show the current user game state
func CurrentCommand(db *sqlx.DB, userID string) slack.ResponseMessage {
	log.Println("Show user current game", userID)
	state, err := datastore.GetUserLastState(db, userID)

	// No state found
	if err != nil {
		return slack.ResponseMessage{
			Text:        "Could not get the current game, but you could `/game start` a new one",
			Attachments: []slack.Attachment{},
		}
	}

	// Get user information
	first, second, err := getUsers(db, state.FirstUserID, state.SecondUserID)
	if err != nil {
		log.Println("Could not get the users information")
	}

	log.Println("Current state ", state, first, second)
	var currentTurn, lastTurn, message string

	if state.TurnID == first.UserID {
		currentTurn = first.Name
		lastTurn = second.Name
	} else {
		currentTurn = second.Name
		lastTurn = first.Name
	}

	if state.Mode == "Turn" {
		message = fmt.Sprintf("It's now *@%s's* turn, last turn was by *@%s* - %s",
			currentTurn, lastTurn, state.Created.Format("15:04:05 02-01-06"))
	} else {
		message = fmt.Sprintf("Game has ended, state [%s], last turn by *@%s* played with *@%s* - %s",
			state.Mode, currentTurn, lastTurn, state.Created.Format("15:04:05 02-01-06"))
	}

	return slack.ResponseMessage{
		Text: message,
		Attachments: []slack.Attachment{
			slack.Attachment{
				Title: "Last game state", Text: "", Fallback: "",
				ImageURL: fmt.Sprintf("https://gametestslack.localtunnel.me/image/%s", state.StateID),
				Color:    "#764FA5",
			},
		},
	}
}
