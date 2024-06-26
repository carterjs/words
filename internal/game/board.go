package game

import (
	"strings"
)

type Board struct {
	minX, minY, maxX, maxY int
	grid                   [][]rune
	directWords            []Word
	indirectWords          []Word
}

func (board Board) AllWords() []Word {
	return append(board.directWords, board.indirectWords...)
}

func (board Board) DirectWords() []Word {
	return board.directWords
}

func (board Board) IndirectWords() []Word {
	return board.indirectWords
}

func NewBoard(words []Word) (Board, error) {
	board := newEmptyBoard(words)

	// add words into grid
	for i, word := range words {

		// TODO: clarify terms
		var overlapping, adjacent bool

		for i, letter := range word.Value {
			dx, dy := vector(word.Direction)
			dx *= i
			dy *= i

			// translate into the bounding space
			x, y := board.translate(word.X+dx, word.Y+dy)

			// check that overlaps are valid
			if board.grid[x][y] != 0 {
				if board.grid[x][y] != letter {
					return Board{}, WordConflictError{
						X:    x,
						Y:    y,
						Want: letter,
						Got:  board.grid[x][y],
					}
				}

				// there is direct overlap
				overlapping = true
			}

			// set the letter
			board.grid[x][y] = letter

			if overlapping {
				continue
			}

			if newWord, ok := board.connectedWord(board.grid, x, y, word.Direction.Other()); ok {
				adjacent = true
				board.indirectWords = append(board.indirectWords, newWord)
			}
		}

		if i > 0 && !overlapping && !adjacent {
			return Board{}, ErrWordNotConnected
		}
	}

	return board, nil
}

func (board Board) connectedWord(grid [][]rune, x, y int, direction Direction) (Word, bool) {
	dx, dy := vector(direction)

	word := string(grid[x][y])
	startX := x
	startY := y

	// starting as far before as possible
	for {
		if startX-dx < 0 || startY-dy < 0 || grid[startX-dx][startY-dy] == 0 {
			break
		}

		word = string(grid[startX-dx][startY-dy]) + word
		startX -= dx
		startY -= dy
	}

	// now go as far after as possible
	endX := x
	endY := y
	for {
		if endX+dx >= len(grid) || endY+dy >= len(grid[0]) || grid[endX+dx][endY+dy] == 0 {
			break
		}

		word += string(grid[endX+dx][endY+dy])
		endX += dx
		endY += dy
	}

	if len(word) == 1 {
		return Word{}, false
	}

	return Word{
		X:         board.minX + startX,
		Y:         board.minY + startY,
		Direction: direction,
		Value:     word,
	}, true
}

func vector(direction Direction) (int, int) {
	if direction == DirectionHorizontal {
		return 1, 0
	} else {
		return 0, 1
	}
}

func (board Board) translate(x, y int) (int, int) {
	return x - board.minX, y - board.minY
}

func newEmptyBoard(words []Word) Board {
	var minX, minY, maxX, maxY int

	for _, word := range words {
		if word.X < minX {
			minX = word.X
		}
		if word.Y < minY {
			minY = word.Y
		}

		if word.Direction == DirectionHorizontal {
			if word.X+len(word.Value) > maxX {
				maxX = word.X + len(word.Value)
			}
			if word.Y+1 > maxY {
				maxY = word.Y + 1
			}
		} else {
			if word.Y+len(word.Value) > maxY {
				maxY = word.Y + len(word.Value)
			}
			if word.X+1 > maxX {
				maxX = word.X + 1
			}
		}
	}

	return Board{
		minX:        minX,
		minY:        minY,
		maxX:        maxX,
		maxY:        maxY,
		grid:        newGrid(maxX-minX, maxY-minY),
		directWords: words,
	}
}

func newGrid(width, height int) [][]rune {
	grid := make([][]rune, width)
	for i := range grid {
		grid[i] = make([]rune, height)
	}

	return grid
}

func (board Board) String() string {
	if len(board.grid) == 0 {
		return ""
	}

	var sb strings.Builder
	for y := range board.grid[0] {
		for x := range board.grid {
			if board.grid[x][y] == 0 {
				sb.WriteRune('_')
			} else {
				sb.WriteRune(board.grid[x][y])
			}
		}

		sb.WriteRune('\n')
	}

	return sb.String()
}
