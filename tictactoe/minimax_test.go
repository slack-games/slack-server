package tictactoe

import "testing"

func TestWinnerMove(t *testing.T) {
	first := MyPlayer
	second := OpponentPlayer
	game := CreateFromField([3][3]uint8{
		{0, 1, 1},
		{0, 0, 0},
		{0, 0, 0},
	}, first, second, first)

	_, spot := AB(game, 3, uint8(MyPlayer), uint8(MyPlayer), MinInt, MaxInt)
	if spot.X != 0 || spot.Y != 0 {
		t.Error("Spot [0,0] is the next best move for the user", spot)
	}
}

func TestOtherPlayerDefence(t *testing.T) {
	first := MyPlayer
	second := OpponentPlayer
	game := CreateFromField([3][3]uint8{
		{0, 0, 0},
		{0, 0, 0},
		{1, 1, 0},
	}, first, second, second)

	_, spot := AB(game, 3, uint8(OpponentPlayer), uint8(OpponentPlayer), MinInt, MaxInt)
	if spot.X != 2 || spot.Y != 2 {
		t.Error("The [2,2], should stop first player to win", spot)
	}
}

func TestDefendOrWinMove(t *testing.T) {
	first := MyPlayer
	second := OpponentPlayer
	game := CreateFromField([3][3]uint8{
		{2, 2, 0},
		{0, 0, 0},
		{1, 1, 0},
	}, first, second, second)

	_, spot := AB(game, 3, uint8(OpponentPlayer), uint8(OpponentPlayer), MinInt, MaxInt)
	if spot.X != 0 || spot.Y != 2 {
		t.Error("The [0,2], should be the win move", spot)
	}
}

func TestDefendingUserMove(t *testing.T) {
	first := MyPlayer
	second := OpponentPlayer
	game := CreateFromField([3][3]uint8{
		{2, 0, 2},
		{0, 0, 0},
		{1, 0, 0},
	}, first, second, first)

	_, spot := AB(game, 3, uint8(MyPlayer), uint8(MyPlayer), MinInt, MaxInt)
	if spot.X != 0 || spot.Y != 1 {
		t.Error("The [0,2], should be the win move", spot)
	}
}
