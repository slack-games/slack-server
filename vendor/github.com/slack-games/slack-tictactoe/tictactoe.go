package tictactoe

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

const (
	// Height is also same
	Width  = 3
	Height = 3
)

const (
	GameOverState State = 1 << iota
	DrawState
	WinState
	TurnState
	StartState
)

const (
	UnkownPlayer Player = iota
	MyPlayer
	OpponentPlayer
)

type Player uint8

type State int

func (s State) String() string {
	var state string

	switch s {
	case GameOverState:
		state = "GameOver"

	case WinState:
		state = "Win"

	case DrawState:
		state = "Draw"

	case TurnState:
		state = "Turn"

	case StartState:
		state = "Start"

	default:
		state = "Unkown"
	}

	return state
}

type Spot struct {
	X, Y uint8
}

func (s *Spot) ToMove() uint8 {
	return s.Y*3 + s.X
}

type Board struct {
	Field [3][3]uint8
}

type TicTacToe struct {
	Board
	First  Player
	Second Player
	Turn   Player
	State
}

func (t *TicTacToe) Start() {
	t.State = StartState
	t.First = MyPlayer
	t.Second = OpponentPlayer

	// Select random turn
	t.SelectRandomPlayer()

	t.State = TurnState
}

func (t *TicTacToe) MakeTurn(x, y uint8) error {

	if !t.hasFreeSpot() {
		t.State = DrawState
		return fmt.Errorf("No free spot at %d - %d", x, y)
	}

	if t.State == GameOverState || t.State == DrawState || t.State == WinState {
		return errors.New("Game over could not make turn")
	}

	// Spot taken, example first player uses number 1
	// second player uses num 2
	// if spot is empty use 0
	if t.Board.Field[x][y] != 0 {
		return fmt.Errorf("Could not redefine the turn %d - %d", x, y)
	}

	// Make the turn
	symbol := t.getCurrentTurnSymbol()
	t.Board.Field[x][y] = symbol

	if t.HasWinner() {
		t.State = WinState
	} else {
		t.State = TurnState
		// Switch the players
		t.ToggleTurn()
	}

	// No errors
	return nil
}

func (t *TicTacToe) HasWinner() bool {
	if t.InRow(1) || t.InRow(2) {
		return true
	}

	return false
}

func (t *TicTacToe) InRow(current uint8) bool {

	// Check for any winnings
	combinations := [][]uint8{
		// Horisontal lines
		[]uint8{0, 1, 2},
		[]uint8{3, 4, 5},
		[]uint8{6, 7, 8},

		// Vertical lines
		[]uint8{0, 3, 6},
		[]uint8{1, 4, 7},
		[]uint8{2, 5, 8},

		// Diagonals
		[]uint8{0, 4, 8},
		[]uint8{2, 4, 6},
	}

	// Helper function
	checkAt := func(index, symbol uint8) bool {
		x, y := GetXY(index)
		return t.Board.Field[x][y] == symbol
	}

	for _, combination := range combinations {
		if checkAt(combination[0], current) &&
			checkAt(combination[1], current) &&
			checkAt(combination[2], current) {

			return true
		}
	}

	// No matches from previous combinations retrun false
	return false
}

func (t *TicTacToe) hasFreeSpot() bool {
	spots := t.GetFreeSpots()
	return len(spots) > 0
}

func (t *TicTacToe) GetFreeSpots() []Spot {
	var spots []Spot

	Loop(func(x, y uint8) {
		if t.Board.Field[x][y] == 0 {
			spots = append(spots, Spot{uint8(x), uint8(y)})
		}
	})

	return spots
}

func (t *TicTacToe) GetRandomFreeSpot() (Spot, error) {
	rand.Seed(int64(time.Now().Nanosecond()))
	spots := t.GetFreeSpots()
	length := len(spots)

	if length == 0 {
		return Spot{}, fmt.Errorf("No random free spot")
	}

	i := rand.Intn(length)

	return spots[i], nil
}

func (t *TicTacToe) ToggleTurn() Player {
	if t.Turn == MyPlayer {
		t.Turn = OpponentPlayer
	} else {
		t.Turn = MyPlayer
	}

	return t.Turn
}

// Deprecate this function
func (t *TicTacToe) getCurrentTurnSymbol() (symbol uint8) {
	return uint8(t.Turn)
}

func (t *TicTacToe) SelectRandomPlayer() Player {
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))

	if r.Float32() < 0.5 {
		t.Turn = MyPlayer
	} else {
		t.Turn = OpponentPlayer
	}

	return t.Turn
}

func (t *TicTacToe) GetBoardAsString() string {
	var state bytes.Buffer

	// Convert board to string
	Loop(func(x, y uint8) {
		state.WriteString(strconv.Itoa(int(t.Board.Field[x][y])))
	})

	return state.String()
}

func (t TicTacToe) String() string {
	board := ""
	for y := 0; y < Height; y++ {
		for x := 0; x < Width; x++ {
			board += fmt.Sprintf("%d", t.Board.Field[x][y])
		}

		board += "\n"
	}
	return board
}

// Loop helper function to replace the repeating double loop pattern
// The outer and inner loop size is fixed by the board width and height
func Loop(fn func(uint8, uint8)) {
	var x, y uint8

	for y = 0; y < Height; y++ {
		for x = 0; x < Width; x++ {
			fn(x, y)
		}
	}
}

// GetXY convert the index from 0..9 to coordinates X,Y
// Example: the 0 -> [0; 0], 3 - [0; 1]
func GetXY(index uint8) (x, y uint8) {
	x = uint8(index % Width)
	y = uint8(index / Width)
	return
}
