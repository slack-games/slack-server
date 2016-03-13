package tictactoe

import "testing"

func TestStartGame(t *testing.T) {
	my := MyPlayer
	ai := OpponentPlayer
	game := CreateFromField([3][3]uint8{
		{0, 2, 0},
		{0, 0, 0},
		{0, 2, 0},
	}, my, ai, ai)

	if game.HasWinner() == true {
		t.Error("Game should have no winners")
	}

	if err := game.MakeTurn(1, 1); err != nil {
		t.Error("Should be able to make move")
	}

	if game.HasWinner() == false {
		t.Error("Game should have a winner now")
	}

	if game.State != WinState {
		t.Error("Should be in a win state")
	}

	if game.Turn != OpponentPlayer {
		t.Error("Ai should have won")
	}
}

func TestPlayerSwitch(t *testing.T) {
	my := MyPlayer
	ai := OpponentPlayer
	game := &TicTacToe{
		First:  my,
		Second: ai,
		Turn:   my,
		Board:  Board{},
		State:  StartState,
	}

	if err := game.MakeTurn(0, 0); err != nil {
		t.Error("Should be able to make move")
	}
	if game.Board.Field[0][0] != 1 {
		t.Error("In field 0-0 should be first player, got ", game.Board.Field[0][0])
	}
	if err := game.MakeTurn(1, 0); err != nil {
		t.Error("Should be able to make move at 1-0")
	}
	if game.Board.Field[1][0] != 2 {
		t.Error("In field 1-0 should be second player, got ", game.Board.Field[1][0])
	}
}

func TestGameFromZero(t *testing.T) {
	my := MyPlayer
	ai := OpponentPlayer
	game := &TicTacToe{
		First:  my,
		Second: ai,
		Turn:   my,
		Board:  Board{},
		State:  StartState,
	}

	turns := [][]uint8{
		{0, 0},
		{1, 0},
		{0, 1},
		{1, 2},
		{0, 2},
	}

	for _, turn := range turns {
		if err := game.MakeTurn(turn[0], turn[1]); err != nil {
			t.Error(err)
		}
	}

	if game.HasWinner() != true {
		t.Error("Game should have winner")
	}

	if err := game.MakeTurn(1, 2); err == nil {
		t.Error("Should not be able to make another move if the game is over")
	}
}

func TestBoardAsString(t *testing.T) {
	my := MyPlayer
	ai := OpponentPlayer

	game := CreateFromField([3][3]uint8{
		{0, 2, 0},
		{0, 0, 0},
		{0, 2, 0},
	}, my, ai, my)

	if "000202000" != game.GetBoardAsString() {
		t.Error("No correct version of board state")
	}
}

func TestBoardGetFreeSpots(t *testing.T) {
	my := MyPlayer
	ai := OpponentPlayer
	game := CreateFromField([3][3]uint8{
		{0, 2, 2},
		{2, 2, 2},
		{2, 2, 2},
	}, my, ai, ai)

	spots := game.GetFreeSpots()

	if len(spots) != 1 {
		t.Error("There should be only one free spot")
	}

	if spots[0].X != 0 && spots[0].Y != 0 {
		t.Error("Not correct free spot returned")
	}
}

func TestCoordinateConversion(t *testing.T) {
	out := [9][2]uint8{
		{0, 0}, {1, 0}, {2, 0},
		{0, 1}, {1, 1}, {2, 1},
		{0, 2}, {1, 2}, {2, 2},
	}

	for i := 0; i < len(out); i++ {
		x, y := GetXY(uint8(i))

		oX := out[i][0]
		oY := out[i][1]

		if x != oX || y != oY {
			t.Errorf("Not getting correct coordinate back index=%d [%d-%d] != [%d-%d]",
				i, oX, oY, x, y)
		}
	}
}
