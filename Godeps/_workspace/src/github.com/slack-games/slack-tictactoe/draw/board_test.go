package draw

import (
	"testing"

	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/slack-games/slack-tictactoe"
)

func TestDrawGame(t *testing.T) {
	game := &tictactoe.TicTacToe{
		First:  tictactoe.MyPlayer,
		Second: tictactoe.OpponentPlayer,
		Turn:   tictactoe.OpponentPlayer,
		Board: tictactoe.Board{
			Field: [3][3]uint8{
				{2, 2, 2},
				{1, 1, 2},
				{1, 1, 2},
			},
		},
		State: tictactoe.StartState,
	}

	image := Draw(game)
	// Save to file
	draw2dimg.SaveToPngFile("board.png", image)
}
