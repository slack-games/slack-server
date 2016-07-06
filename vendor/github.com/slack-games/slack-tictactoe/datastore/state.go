package datastore

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/slack-games/slack-tictactoe"
)

type State struct {
	StateID      string    `db:"state_id"`
	State        string    `db:"state"`
	TurnID       string    `db:"turn"`
	Mode         string    `db:"mode"`
	FirstUserID  string    `db:"first_user_id"`
	SecondUserID string    `db:"second_user_id"`
	ParentID     string    `db:"parent_state_id"`
	Created      time.Time `db:"created_at"`
}

func (s State) String() string {
	return fmt.Sprintf("#[%s] - %s %s %s %s %s %s",
		s.StateID, s.State, s.TurnID, s.Mode, s.FirstUserID, s.SecondUserID, s.Created)
}

func CreateStateFromBoard(game *tictactoe.TicTacToe, state State) *State {

	return &State{
		State:        game.GetBoardAsString(),
		TurnID:       state.TurnID,
		Mode:         fmt.Sprintf("%s", game.State),
		FirstUserID:  state.FirstUserID,
		SecondUserID: state.SecondUserID,
		ParentID:     state.StateID,
		Created:      time.Now(),
	}
}

func CreateTicTacToeBoard(state State) *tictactoe.TicTacToe {
	turn := tictactoe.MyPlayer

	if state.TurnID == state.SecondUserID {
		turn = tictactoe.OpponentPlayer
	}

	field := [3][3]uint8{}
	// Fill the board
	cells := strings.Split(state.State, "")

	for i := 0; i < len(cells); i++ {
		x, y := tictactoe.GetXY(uint8(i))
		if num, err := strconv.ParseInt(cells[i], 10, 8); err == nil {
			field[x][y] = uint8(num)
		}
	}

	game := &tictactoe.TicTacToe{
		First:  tictactoe.MyPlayer,
		Second: tictactoe.OpponentPlayer,
		Turn:   turn,
		Board: tictactoe.Board{
			Field: field,
		},
		State: tictactoe.StartState,
	}
	return game
}

func GetState(db *sqlx.DB, id string) (State, error) {
	state := State{}

	// TODO: switch from * to field names
	err := db.Get(&state, `SELECT * FROM ttt.states WHERE state_id=$1 LIMIT 1`, id)
	return state, err
}

func GetUserLastState(db *sqlx.DB, id string) (State, error) {
	state := State{}

	query := `
		SELECT *
		FROM ttt.states
		WHERE
			first_user_id=$1 OR second_user_id=$1
		ORDER BY created_at DESC LIMIT 1;
	`

	err := db.Get(&state, query, id)
	return state, err
}

func NewState(db *sqlx.DB, state State) (string, error) {
	sql := `
		INSERT INTO ttt.states
			(state, turn, mode, first_user_id, second_user_id, parent_state_id)
		VALUES
			(:state, :turn, :mode, :first_user_id, :second_user_id, :parent_state_id)
		RETURNING state_id
	`
	var id string

	rows, err := db.NamedQuery(sql, state)
	if err != nil {
		return id, err
	}

	if rows.Next() {
		rows.Scan(&id)
	}
	return id, err
}
