package draw

import (
	"image"
	"image/color"
	"math"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
	kit "github.com/llgcode/draw2d/draw2dkit"
	"github.com/slack-games/slack-tictactoe"
)

const (
	Width  = 350
	Height = 350
	Offset = 25.0
)

var DefaultColor color.RGBA
var FirstColor color.RGBA
var SecondColor color.RGBA

func init() {
	// #444444
	DefaultColor = color.RGBA{0x44, 0x44, 0x44, 0xff}
	// #0A63BB
	FirstColor = color.RGBA{0x0a, 0x63, 0xbb, 0xff}
	// #6D083E
	SecondColor = color.RGBA{0x6d, 0x08, 0x3e, 0xff}
}

func DrawLines(gc *draw2dimg.GraphicContext) {
	// Draw vertical lines
	gc.MoveTo(Offset+100, Offset)
	gc.LineTo(Offset+100, Height-Offset)
	gc.Close()
	gc.FillStroke()

	gc.MoveTo(Offset+200, Offset)
	gc.LineTo(Offset+200, Height-Offset)
	gc.Close()
	gc.FillStroke()

	// Draw horisontal lines
	gc.MoveTo(Offset, Offset+100)
	gc.LineTo(Width-Offset, Offset+100)
	gc.Close()
	gc.FillStroke()

	gc.MoveTo(Offset, Offset+200)
	gc.LineTo(Width-Offset, Offset+200)
	gc.Close()
	gc.FillStroke()
}

func DrawCross(gc *draw2dimg.GraphicContext, x, y float64, size float64) {

	a := size / math.Sqrt(2)
	hA := a / 2

	gc.MoveTo(x-hA, y-hA)
	gc.LineTo(x+hA, y+hA)
	gc.Close()
	gc.FillStroke()

	gc.MoveTo(x+hA, y-hA)
	gc.LineTo(x-hA, y+hA)
	gc.Close()
	gc.FillStroke()
}

func DrawSpot(gc *draw2dimg.GraphicContext, board tictactoe.Board) {
	cellSize := 100.0
	hCell := cellSize / 2
	index := 0

	for y := 0; y < len(board.Field); y++ {
		for x := 0; x < len(board.Field[0]); x++ {
			xPos := float64(x)*cellSize + Offset + hCell
			yPos := float64(y)*cellSize + Offset + hCell

			gc.SetLineWidth(5)
			if board.Field[x][y] == 1 {
				gc.SetStrokeColor(FirstColor)
				kit.Circle(gc, xPos, yPos, 30)
				gc.FillStroke()
			}

			if board.Field[x][y] == 2 {
				gc.SetStrokeColor(SecondColor)
				DrawCross(gc, xPos, yPos, 60)
				gc.FillStroke()
			}

			// Draw spot number
			// gc.SetFillColor(image.Black)
			// gc.SetLineWidth(1)
			// gc.StrokeStringAt(fmt.Sprintf("%d", index), xPos+hCell-15, yPos+hCell-15)
			// gc.FillStroke()
			index++
		}
	}
	gc.SetStrokeColor(DefaultColor)
}

func DrawWinLines(gc *draw2dimg.GraphicContext, board tictactoe.Board, symbol uint8) {
	cellSize := 100.0
	hCell := cellSize / 2
	redColor := color.RGBA{0xFF, 0x0, 0x0, 0xFF}

	gc.SetLineCap(draw2d.RoundCap)
	gc.SetLineJoin(draw2d.RoundJoin)
	gc.SetStrokeColor(redColor)
	gc.SetLineWidth(7)

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

	toXY := func(index uint8) (x, y uint8) {
		x = index / 3
		y = index % 3
		return
	}

	// Helper function
	checkAt := func(index, symbol uint8) bool {
		x, y := toXY(index)
		return board.Field[x][y] == symbol
	}

	for _, combination := range combinations {
		if checkAt(combination[0], symbol) &&
			checkAt(combination[1], symbol) &&
			checkAt(combination[2], symbol) {

			x, y := toXY(combination[0])

			// Draw line for combination
			x1 := float64(x)*cellSize + Offset + hCell
			y1 := float64(y)*cellSize + Offset + hCell

			x, y = toXY(combination[2])

			x2 := float64(x)*cellSize + Offset + hCell
			y2 := float64(y)*cellSize + Offset + hCell

			gc.MoveTo(x1, y1)
			gc.LineTo(x2, y2)
			gc.Close()
			gc.Stroke()
		}
	}
}

func Draw(game *tictactoe.TicTacToe) image.Image {

	// Initialize the graphic context on an RGBA image
	dest := image.NewRGBA(image.Rect(0, 0, Width, Height))
	gc := draw2dimg.NewGraphicContext(dest)

	// Draw letters
	// draw2d.SetFontFolder(fontPath)
	//
	// gc.SetFontData(draw2d.FontData{
	// 	Name:   "luxi",
	// 	Family: draw2d.FontFamilyMono,
	// 	Style:  draw2d.FontStyleNormal,
	// })
	// gc.SetFontSize(9)

	// Set some properties
	gc.SetFillColor(color.Transparent)
	gc.SetStrokeColor(color.RGBA{0x44, 0x44, 0x44, 0xff})
	gc.SetLineWidth(5)

	// Horisontal and vertical lines
	DrawLines(gc)

	// Draw spot at
	DrawSpot(gc, game.Board)

	// DrawWinLines
	DrawWinLines(gc, game.Board, 1)
	DrawWinLines(gc, game.Board, 2)

	return dest
}
