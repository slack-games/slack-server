package commands

import (
	"errors"
	"image"

	"github.com/jmoiron/sqlx"
	tttdatastore "github.com/slack-games/slack-tictactoe/datastore"
	drawBoard "github.com/slack-games/slack-tictactoe/draw"
)

// GetGameImage returns the image by state
func GetGameImage(db *sqlx.DB, stateID string) (image.Image, error) {
	state, err := tttdatastore.GetState(db, stateID)
	if err != nil {
		return nil, errors.New("Could not get the state")
	}

	ttt := tttdatastore.CreateTicTacToeBoard(state)

	return drawBoard.Draw(ttt), nil
}
