package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/slack-games/slack-client"
	"github.com/slack-games/slack-tictactoe/datastore"
)

// CurrentCommand show the current user game state
func CurrentCommand(db *sqlx.DB, userID string) slack.ResponseMessage {
	baseURL := os.Getenv("BASE_PATH")

	log.Println("Show user current game", userID)
	state, err := datastore.GetUserLastState(db, userID)

	// No state found
	if err != nil {
		return slack.TextOnly("Could not get the current game, but you could `/ttt start` a new one")
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

	if state.Mode == "Turn" || state.Mode == "Start" {
		message = fmt.Sprintf("It's now *@%s's* [%s] turn, last turn was by *@%s* - _at %s_",
			currentTurn, getSymbol(state, userID), lastTurn, state.Created.Format("15:04:05 02-01-06"))
	} else {
		message = fmt.Sprintf(":tada: Game won by *@%s*, played with *@%s* - _at %s_. For a new game `/ttt start` :tada:",
			currentTurn, lastTurn, state.Created.Format("15:04:05 02-01-06"))
	}

	return slack.ResponseMessage{
		Text: message,
		Attachments: []slack.Attachment{
			slack.Attachment{
				Title:    "Last game state",
				ImageURL: fmt.Sprintf("%s/game/tictactoe/image/%s", baseURL, state.StateID),
				Color:    "#764FA5",
			},
		},
	}
}
