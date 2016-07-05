package commands

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/slack-games/slack-client"
	"github.com/slack-games/slack-tictactoe"
	tttdatastore "github.com/slack-games/slack-tictactoe/datastore"
)

// StartCommand is command to start
func StartCommand(db *sqlx.DB, userID string) slack.ResponseMessage {
	var attachment slack.Attachment
	baseURL := os.Getenv("BASE_PATH")
	message := "There's already existing a game, you have to finish it before starting a new"

	// Try to get user last state
	state, err := tttdatastore.GetUserLastState(db, userID)

	if err != nil {
		// No state found
		if err == sql.ErrNoRows {
			stateID, newState := createNewState(db, userID)
			symbol := getSymbol(newState, userID)

			message = fmt.Sprintf("Created a new clean game state, your turn as %s", symbol)
			attachment = slack.Attachment{
				Title:    "Last game state",
				Fallback: "Text fallback if image fails",
				ImageURL: fmt.Sprintf("%s/game/tictactoe/image/%s", baseURL, stateID),
				Color:    "#764FA5",
			}

			log.Println("New state id", stateID)
		} else {
			log.Println("Error could not get the user state")
		}
	} else if isGameOver(state) {
		stateID, newState := createNewState(db, userID)
		symbol := getSymbol(newState, userID)

		message = fmt.Sprintf("Created a new game state, your turn as %s. To make move `/ttt move [1-9]`.",
			symbol)
		attachment = slack.Attachment{
			Title:    "New game state",
			Fallback: "Text fallback if image fails",
			ImageURL: fmt.Sprintf("%s/game/tictactoe/image/%s", baseURL, stateID),
			Color:    "#764FA5",
		}
	} else {
		attachment = slack.Attachment{
			Title:    "Last game state",
			Fallback: "Text fallback if image fails",
			ImageURL: fmt.Sprintf("%s/game/tictactoe/image/%s", baseURL, state.StateID),
			Color:    "#764FA5",
		}
	}

	return slack.ResponseMessage{
		Text:        message,
		Attachments: []slack.Attachment{attachment},
	}
}

func getSymbol(state tttdatastore.State, userID string) string {
	if state.FirstUserID == userID {
		return ":o:"
	}
	return ":x:"
}

func createNewState(db *sqlx.DB, userID string) (ID string, state tttdatastore.State) {
	now := time.Now().Unix()

	state = tttdatastore.State{
		State:        "000000000",
		TurnID:       userID,
		Mode:         "Start",
		FirstUserID:  "U000000000",
		SecondUserID: userID,
		ParentID:     "00000000-0000-0000-0000-000000000000",
		Created:      time.Now(),
	}

	if (now % 2) == 0 {
		state.FirstUserID = userID
		state.SecondUserID = "U000000000"
		state.State = "000020000"
	}

	log.Println("Create a new state")
	ID, err := tttdatastore.NewState(db, state)
	if err != nil {
		log.Fatalln("Could not create a new state", err)
	}
	return
}

func isGameOver(state tttdatastore.State) bool {
	return state.Mode == fmt.Sprintf("%s", tictactoe.GameOverState) ||
		state.Mode == fmt.Sprintf("%s", tictactoe.WinState) ||
		state.Mode == fmt.Sprintf("%s", tictactoe.DrawState)
}
