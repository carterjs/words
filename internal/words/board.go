package words

import (
	"fmt"
	"strconv"
	"strings"
)

type (
	Board struct {
		GameID                 string
		Grid                   map[Point]rune
		Words                  []Word
		MinX, MinY, MaxX, MaxY int
		Config                 Config
	}

	Point string
)

func NewPoint(x, y int) Point {
	return Point(fmt.Sprint(x, ",", y))
}

func (p Point) X() int {
	s, _, _ := strings.Cut(string(p), ",")
	x, _ := strconv.Atoi(s)
	return x
}

func (p Point) Y() int {
	_, s, _ := strings.Cut(string(p), ",")
	y, _ := strconv.Atoi(s)
	return y
}

func newBoard(gameID string, config Config) *Board {
	board := &Board{
		GameID: gameID,
		Grid:   make(map[Point]rune),
		Config: config,
	}

	return board
}

func (board *Board) tryWordPlacement(w Word) (PlacementResult, error) {
	needsConnection := true
	if len(board.Grid) == 0 {
		if w.X != 0 || w.Y != 0 {
			return PlacementResult{}, ErrFirstWordNotCentered
		}

		needsConnection = false
	}

	result := PlacementResult{
		DirectWord: w,
	}

	// remember what the blank is being used for
	blankLetterMapping := make(map[int]rune)
	for i := range w.Blanks {
		_, _, letter, _ := w.Index(i)
		blankLetterMapping[i] = letter
	}

	for i := range w.Letters {
		x, y, letter, _ := w.Index(i)

		// see if there's already a letter there
		if currentLetter, isSet := board.GetLetter(x, y); isSet {
			if currentLetter != letter {
				return PlacementResult{}, WordConflictError{
					X:    x,
					Y:    y,
					Want: letter,
					Got:  currentLetter,
				}
			}

			// assert that the word has a blank
			// this should be a no-op if the word is already marked as blank but is necessary for scoring
			if _, isBlank := blankLetterMapping[i]; isBlank {
				result.DirectWord = result.DirectWord.WithBlank(i)
			}

			// successful overlap with an existing word
			// can't have indirect words if it's already overlapping
			needsConnection = false
			continue
		}

		// check for modifier
		if modifier, hasModifier := board.getModifier(x, y); hasModifier {
			if result.Modifiers == nil {
				result.Modifiers = make(map[int]Modifier)
			}

			result.Modifiers[i] = modifier
		}

		// track used letter (or blank)
		if _, isBlank := blankLetterMapping[i]; isBlank {
			result.LettersUsed = append(result.LettersUsed, BlankLetter)
		} else {
			result.LettersUsed = append(result.LettersUsed, letter)
		}

		// look for indirect words formed by this placement
		if indirectWord, hasIndirectWord := board.wordFormedByNewLetter(letter, x, y, w.Direction.Other()); hasIndirectWord {
			result.IndirectWords = append(result.IndirectWords, indirectWord)
			needsConnection = false
		}
	}

	if needsConnection {
		return PlacementResult{}, ErrWordNotConnected
	}

	if len(result.LettersUsed) == 0 {
		return PlacementResult{}, ErrUnchanged
	}

	return result.withComputedPoints(board.Config.LetterPoints), nil
}

func (board *Board) getModifier(x, y int) (Modifier, bool) {
	return board.Config.Modifiers.Get(x, y)
}

func (board *Board) GetLetter(x, y int) (rune, bool) {
	letter, exists := board.Grid[NewPoint(x, y)]
	return letter, exists
}

func (board *Board) GetModifier(x, y int) (Modifier, bool) {
	return board.Config.Modifiers.Get(x, y)
}

func (board *Board) placeWord(w Word) (PlacementResult, error) {
	wordPlacement, err := board.tryWordPlacement(w)
	if err != nil {
		return PlacementResult{}, err
	}
	//board.indirectWords = append(board.indirectWords, wordPlacement.IndirectWords...)

	for i := range w.Letters {
		x, y, letter, _ := w.Index(i)
		board.set(x, y, letter)
	}
	//board.directWords = append(board.directWords, w)

	board.MinX = min(board.MinX, w.X)
	board.MinY = min(board.MinY, w.Y)

	if w.Direction == DirectionHorizontal {
		board.MaxX = max(board.MaxX, w.X+len(w.Letters))
		board.MaxY = max(board.MaxY, w.Y+1)
	} else {
		board.MaxY = max(board.MaxY, w.Y+len(w.Letters))
		board.MaxX = max(board.MaxX, w.X+1)
	}

	board.Words = append(board.Words, w)

	return wordPlacement, nil
}

func (board *Board) removeLastWord() error {
	if len(board.Words) == 0 {
		return ErrNothingToUndo
	}

	board.MinX = 0
	board.MinY = 0
	board.MaxX = 0
	board.MaxY = 0
	board.Grid = make(map[Point]rune)
	words := board.Words[:len(board.Words)-1]
	board.Words = []Word{}

	for _, word := range words {
		_, _ = board.placeWord(word)
	}

	return nil
}

func (board *Board) set(x, y int, letter rune) {
	board.Grid[NewPoint(x, y)] = letter
}

func (board *Board) wordFormedByNewLetter(letter rune, x, y int, direction Direction) (Word, bool) {
	dx, dy := direction.Vector(1)

	s := string(letter)
	var blanks []int
	startX := x
	startY := y

	// starting as far before as possible
	for {
		letter, isSet := board.GetLetter(startX-dx, startY-dy)
		if !isSet {
			break
		}

		s = string(letter) + s
		startX -= dx
		startY -= dy
	}

	// now go as far after as possible
	endX := x
	endY := y
	for {
		letter, isSet := board.GetLetter(endX+dx, endY+dy)
		if !isSet {
			break
		}

		s += string(letter)
		endX += dx
		endY += dy
		if letter == BlankLetter {
			blanks = append(blanks, len(s)-1)
		}
	}

	if len(s) == 1 {
		return Word{}, false
	}

	newWord := NewWord(startX, startY, direction, s)

	for _, i := range blanks {
		newWord = newWord.WithBlank(i)
	}

	return newWord, true
}

func (board *Board) String() string {
	board.Config = StandardConfig
	grid := make([][]rune, board.MaxY-board.MinY)
	for i := range grid {
		grid[i] = make([]rune, board.MaxX-board.MinX)
	}

	var sb strings.Builder
	for point, letter := range board.Grid {
		grid[point.Y()-board.MinY][point.X()-board.MinX] = letter
	}

	buffer := 1
	width := 3
	minX := -8
	maxX := 8
	minY := -8
	maxY := 8

	if board.MinX <= minX {
		minX = board.MinX - buffer
	}

	if board.MaxX >= maxX {
		maxX = board.MaxX - 1 + buffer
	}

	if board.MinY <= minY {
		minY = board.MinY - buffer
	}

	if board.MaxY >= maxY {
		maxY = board.MaxY - 1 + buffer
	}

	// heading label
	sb.WriteString(centered(" ", width+1))
	for x := minX; x <= maxX; x++ {
		sb.WriteString(centered(fmt.Sprint(x), width))
		sb.WriteRune(' ')
	}
	sb.WriteRune('\n')

	for y := minY; y <= maxY; y++ {
		sb.WriteString(centered(fmt.Sprint(y), width))
		sb.WriteRune('\u2502')
		for x := minX; x <= maxX; x++ {
			var letter rune
			if x >= board.MinX && x < board.MaxX && y >= board.MinY && y < board.MaxY {
				letter = grid[y-board.MinY][x-board.MinX]
			}

			if letter == 0 {
				if modifier, hasModifier := board.Config.Modifiers.Get(x, y); hasModifier {
					sb.WriteString(centered(string(modifier), width))
				} else if x == 0 && y == 0 {
					sb.WriteString(filled("\u2592", width))
				} else {
					sb.WriteString(filled("\u2591", width))
				}
			} else {
				sb.WriteString(centered(string(letter), width))
			}
			sb.WriteRune('\u2502')
		}
		sb.WriteString(centered(fmt.Sprint(y), width))
		sb.WriteRune('\n')
	}

	sb.WriteString(centered(" ", width+1))
	for x := minX; x <= maxX; x++ {
		sb.WriteString(centered(fmt.Sprint(x), width))
		sb.WriteRune(' ')
	}

	return sb.String()
}

func centered(s string, width int) string {
	return fmt.Sprintf("%*s", -width, fmt.Sprintf("%*s", (width+len(s))/2, s))
}

func filled(s string, width int) string {
	return strings.Repeat(s, width)
}
