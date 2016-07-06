package tictactoe

const (
	// MaxInt maximum field value
	MaxInt = 0xFF
	// MinInt minimum field value
	MinInt = -0xFF
)

func GetNewGameState(oldGame TicTacToe, x, y uint8) TicTacToe {
	newGame := oldGame
	newGame.MakeTurn(x, y)

	return newGame
}

func switchPlayer(player uint8) uint8 {
	if player == uint8(MyPlayer) {
		return uint8(OpponentPlayer)
	}
	return uint8(MyPlayer)
}

func evaluateBoard(game TicTacToe, maximizer uint8) int {
	// Check for the maximizer
	if game.InRow(maximizer) {
		return 10
	}

	// Check the other player
	if game.InRow(switchPlayer(maximizer)) {
		return -10
	}

	// Check for the middle field
	if game.Field[1][1] == maximizer {
		return 2
	}

	return -2
}

func AB(game TicTacToe, depth, maximizer, player uint8, a, b int) (score int, spot Spot) {
	moves := game.GetFreeSpots()

	// No moves available, or depth reached
	if len(moves) <= 0 || depth <= 0 {
		// Evaluate function from the opponent and maximizer view point

		score = evaluateBoard(game, maximizer)
		spot = Spot{0xFF, 0xFF}
		return
	}

	for _, move := range moves {
		newGame := GetNewGameState(game, move.X, move.Y)

		if maximizer == player {
			// Alpha
			score, _ = AB(newGame, depth-1, maximizer, switchPlayer(player), a, b)
			if score > a {
				a = score
				spot = move
			}
		} else {
			// Beta
			score, _ = AB(newGame, depth-1, maximizer, switchPlayer(player), a, b)
			if score < b {
				b = score
				spot = move
			}
		}

		if a >= b {
			break
		}
	}

	if player == maximizer {
		score = a
	} else {
		score = b
	}
	return
}
