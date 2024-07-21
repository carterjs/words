package game

import (
	"strings"
)

type Board struct {
	space                  map[point]rune
	minX, minY, maxX, maxY int
	directWords            []Word
	indirectWords          []Word
}

type point struct {
	x, y int
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

func NewBoard(words []Word) (*Board, error) {
	board := &Board{
		space: make(map[point]rune),
	}

	// add words into grid
	for _, word := range words {
		_, err := board.PlaceWord(word)
		if err != nil {
			return nil, err
		}
	}

	return board, nil
}

type PlacementResult struct {
	LettersUsed   []rune
	IndirectWords []Word
}

func (board Board) TryWordPlacement(word Word) (PlacementResult, error) {
	var connected bool
	var lettersRequired []rune
	var indirectWords []Word

	for i := range word.Letters {
		x, y, letter := word.Get(i)

		if currentLetter, isSet := board.get(x, y); isSet {
			if currentLetter != letter {
				return PlacementResult{}, WordConflictError{
					X:    x,
					Y:    y,
					Want: letter,
					Got:  currentLetter,
				}
			}

			// can't have indirect words if it's already overlapping
			connected = true
			continue
		}

		lettersRequired = append(lettersRequired, letter)

		if indirectWord, hasIndirectWord := board.wordFormedByNewLetter(letter, x, y, word.Direction.Other()); hasIndirectWord {
			indirectWords = append(indirectWords, indirectWord)
			connected = true
		}
	}

	if !connected && len(board.directWords) > 0 {
		return PlacementResult{}, ErrWordNotConnected
	}

	if len(lettersRequired) == 0 {
		return PlacementResult{}, ErrUnchangedBoard
	}

	return PlacementResult{
		LettersUsed:   lettersRequired,
		IndirectWords: indirectWords,
	}, nil
}

func (board Board) get(x, y int) (rune, bool) {
	letter, exists := board.space[point{x, y}]
	return letter, exists
}

func (board *Board) PlaceWord(word Word) (PlacementResult, error) {
	wordPlacement, err := board.TryWordPlacement(word)
	if err != nil {
		return PlacementResult{}, err
	}
	board.indirectWords = append(board.indirectWords, wordPlacement.IndirectWords...)

	for i := range word.Letters {
		board.set(word.Get(i))
	}
	board.directWords = append(board.directWords, word)

	board.minX = min(board.minX, word.X)
	board.minY = min(board.minY, word.Y)

	if word.Direction == DirectionHorizontal {
		board.maxX = max(board.maxX, word.X+len(word.Letters))
		board.maxY = max(board.maxY, word.Y+1)
	} else {
		board.maxY = max(board.maxY, word.Y+len(word.Letters))
		board.maxX = max(board.maxX, word.X+1)
	}

	return wordPlacement, nil
}

func (board *Board) set(x, y int, letter rune) {
	board.space[point{x, y}] = letter
}

func (board Board) wordFormedByNewLetter(letter rune, x, y int, direction Direction) (Word, bool) {
	dx, dy := direction.Vector(1)

	word := string(letter)
	startX := x
	startY := y

	// starting as far before as possible
	for {
		letter, isSet := board.get(startX-dx, startY-dy)
		if !isSet {
			break
		}

		word = string(letter) + word
		startX -= dx
		startY -= dy
	}

	// now go as far after as possible
	endX := x
	endY := y
	for {
		letter, isSet := board.get(endX+dx, endY+dy)
		if !isSet {
			break
		}

		word += string(letter)
		endX += dx
		endY += dy
	}

	if len(word) == 1 {
		return Word{}, false
	}

	return Word{
		X:         startX,
		Y:         startY,
		Direction: direction,
		Letters:   []rune(word),
	}, true
}

func (board Board) String() string {
	if len(board.space) == 0 {
		return "(no words)"
	}

	grid := make([][]rune, board.maxX-board.minX)
	for i := range grid {
		grid[i] = make([]rune, board.maxY-board.minY)
	}

	var sb strings.Builder
	for point, letter := range board.space {
		grid[point.x-board.minX][point.y-board.minY] = letter
	}

	for _, row := range grid {
		for _, letter := range row {
			if letter == 0 {
				sb.WriteRune('_')
			} else {
				sb.WriteRune(letter)
			}
		}
		sb.WriteRune('\n')
	}

	return sb.String()
}
