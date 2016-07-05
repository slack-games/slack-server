package commands

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/slack-games/slack-client"
	"github.com/slack-games/slack-server/datastore"
	"github.com/slack-games/slack-tictactoe"
	tttdatastore "github.com/slack-games/slack-tictactoe/datastore"
)

const (
	oSymbol = ":o:"
	xSymbol = ":x:"
)

// MoveCommand defines the tic tac toe moves
func MoveCommand(db *sqlx.DB, userID string, spot uint8) slack.ResponseMessage {
	baseURL := os.Getenv("BASE_PATH")
	state, err := tttdatastore.GetUserLastState(db, userID)

	if err != nil {
		// No state found
		if err == sql.ErrNoRows {
			return slack.TextOnly("You can not make any moves before the game has started `/ttt start`")
		}
	}

	// Check the game states
	if isGameOver(state) {
		log.Println("Game is already over")
		return slack.TextOnly("Current game is over, but you can always start a new game `/ttt start`")
	}

	// Convert 0-9 into x-y point
	x, y := tictactoe.GetXY(spot)

	game := tttdatastore.CreateTicTacToeBoard(state)

	log.Println("Should be able to make move", x, y)
	err = game.MakeTurn(x, y)
	if err != nil {
		return slack.TextOnly(fmt.Sprintf("Could not make the move to %d :scream_cat:", spot))
	}

	freeSpot, err := game.GetRandomFreeSpot()
	if err != nil {
		log.Println("No free spot where to move")
	}

	if err = game.MakeTurn(freeSpot.X, freeSpot.Y); err != nil {
		log.Println("Should be able to make move", freeSpot)
	}

	newState := tttdatastore.CreateStateFromBoard(game, state)
	stateID, err := tttdatastore.NewState(db, *newState)
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

	userSymbol := xSymbol
	opponentSymbol := oSymbol

	if userID == first.UserID {
		userSymbol = oSymbol
		opponentSymbol = xSymbol
	}

	return slack.ResponseMessage{
		Text: fmt.Sprintf(":space_invader: You (%s) made move to *[%d]*, opponent (%s) made next move to *[%d]*, state *'%s'*",
			userSymbol, spot, opponentSymbol, freeSpot.ToMove(), newState.Mode),
		Attachments: []slack.Attachment{
			slack.Attachment{
				Title:    "The current game state",
				ImageURL: fmt.Sprintf("%s/game/tictactoe/image/%s", baseURL, stateID),
				Color:    "#764FA5",
			},
		},
	}
}

func getUsers(db *sqlx.DB, firstID, secondID string) (first datastore.User, second datastore.User, err error) {
	first, err = datastore.GetUser(db, firstID)
	second, err = datastore.GetUser(db, secondID)
	return
}
