package commands

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/riston/slack-server/datastore"
	"github.com/riston/slack-server/slack"
	"github.com/riston/slack-server/tictactoe"
)

func MoveCommand(db *sqlx.DB, userID string, spot uint8) slack.ResponseMessage {
	state, err := datastore.GetUserLastState(db, userID)

	if err != nil {
		// No state found
		if err == sql.ErrNoRows {
			return slack.ResponseMessage{
				Text:        "You can not make any moves before the game has started `/game start`",
				Attachments: []slack.Attachment{},
			}
		}
	}

	// Check the game states
	if isGameOver(state) {
		log.Println("Game is already over")
		return slack.ResponseMessage{
			Text:        "Current game is over, but you can always start a new game `/game start`",
			Attachments: []slack.Attachment{},
		}
	}

	// Convert 0-9 into x-y point
	x, y := tictactoe.GetXY(spot)

	game := datastore.CreateTicTacToeBoard(state)

	if err := game.MakeTurn(x, y); err != nil {
		log.Println("Should be able to make move", x, y)
	}

	freeSpot, err := game.GetRandomFreeSpot()
	if err != nil {
		log.Println("No free spot where to move")
	}

	if err := game.MakeTurn(freeSpot.X, freeSpot.Y); err != nil {
		log.Println("Should be able to make move", freeSpot)
	}

	newState := datastore.CreateStateFromBoard(game, state)
	stateID, err := datastore.NewState(db, *newState)
	if err != nil {
		log.Println("Could not save the new state", err)
	}

	// Get user information
	first, second, err := getUsers(db, newState.FirstUserID, newState.SecondUserID)
	if err != nil {
		log.Println("Could not get the users information")
	}

	fmt.Println("Users ", first, second)
	fmt.Println("Matching the test ", spot, stateID)

	return slack.ResponseMessage{
		fmt.Sprintf("You made move to [%d], opponent made next move to [%d], state %s", spot, freeSpot, newState.Mode),
		[]slack.Attachment{
			slack.Attachment{
				"The current game state", "", "",
				fmt.Sprintf("https://gametestslack.localtunnel.me/image/%s", stateID),
				"#764FA5",
			},
		},
	}
}

func getUsers(db *sqlx.DB, firstID, secondID string) (first datastore.User, second datastore.User, err error) {
	first, err = datastore.GetUser(db, firstID)
	second, err = datastore.GetUser(db, secondID)
	return
}
