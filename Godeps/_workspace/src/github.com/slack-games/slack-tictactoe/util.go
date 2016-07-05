package tictactoe

// CreateFromField is a method to quickly create TicTacToe structure
func CreateFromField(from [3][3]uint8, p1, p2, turn Player) (game TicTacToe) {
	game = TicTacToe{
		First:  p1,
		Second: p2,
		Turn:   turn,
		Board: Board{
			Field: from,
		},
		State: StartState,
	}
	return
}
