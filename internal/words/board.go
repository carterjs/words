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
		Blanks                 map[Point]struct{}
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

func (p Point) Offset(dx, dy int) Point {
	return NewPoint(p.X()+dx, p.Y()+dy)
}

func NewBoard(gameID string, config Config) *Board {
	board := &Board{
		GameID: gameID,
		Grid:   make(map[Point]rune),
		Blanks: make(map[Point]struct{}),
		Config: config,
	}

	return board
}

func (board *Board) tryWordPlacement(w Word) (PlacementResult, error) {
	needsConnection := true
	if len(board.Grid) == 0 {
		// if word does not intersect center
		_, intersectsCenter := w.Get(NewPoint(0, 0))
		if !intersectsCenter {
			return PlacementResult{}, ErrFirstWordNotCentered
		}

		needsConnection = false
	}

	result := PlacementResult{
		LettersUsed: make(map[Point]rune),
		DirectWord:  w,
	}

	for i := range w.Letters {
		point, letter, _ := w.Index(i)

		// see if there's already a letter there
		if currentLetter, isSet := board.GetLetter(point); isSet {
			if currentLetter != letter {
				return PlacementResult{}, WordConflictError{
					X:    point.X(),
					Y:    point.Y(),
					Want: letter,
					Got:  currentLetter,
				}
			}

			// assert that the word has a blank
			// this should be a no-op if the word is already marked as blank but is necessary for scoring
			if _, isBlank := board.Blanks[point]; isBlank {
				result.DirectWord = result.DirectWord.WithBlanks(point)
			}

			// successful overlap with an existing word
			// can't have indirect words if it's already overlapping
			needsConnection = false
			continue
		}

		// check for modifier
		if modifier, hasModifier := board.getModifier(point); hasModifier {
			if result.Modifiers == nil {
				result.Modifiers = make(map[int]Modifier)
			}

			result.Modifiers[i] = modifier
		}

		// track used letter (or blank)
		if _, isBlank := w.Blanks[point]; isBlank {
			result.LettersUsed[point] = BlankLetter
		} else {
			result.LettersUsed[point] = letter
		}

		// look for indirect words formed by this placement
		if indirectWord, hasIndirectWord := board.wordFormedByNewLetter(letter, point, w.Direction.Other()); hasIndirectWord {
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

	// assert no letters before or after
	before := w.Start.Offset(w.Direction.Vector(-1))
	after := w.Start.Offset(w.Direction.Vector(len(w.Letters)))

	// before word
	if _, isSet := board.GetLetter(before); isSet {
		return PlacementResult{}, ErrIncomplete
	}

	if _, isSet := board.GetLetter(after); isSet {
		return PlacementResult{}, ErrIncomplete
	}

	return result.withComputedPoints(board.Config.LetterPoints), nil
}

func (board *Board) getModifier(point Point) (Modifier, bool) {
	return board.Config.Modifiers.Get(point.X(), point.Y())
}

func (board *Board) GetLetter(point Point) (rune, bool) {
	letter, exists := board.Grid[point]
	return letter, exists
}

func (board *Board) GetModifier(x, y int) (Modifier, bool) {
	return board.Config.Modifiers.Get(x, y)
}

func (board *Board) PlaceWord(w Word) (PlacementResult, error) {
	wordPlacement, err := board.tryWordPlacement(w)
	if err != nil {
		return PlacementResult{}, err
	}
	//board.indirectWords = append(board.indirectWords, wordPlacement.IndirectWords...)

	for i := range w.Letters {
		point, letter, _ := w.Index(i)
		board.set(point, letter)
	}
	//board.directWords = append(board.directWords, w)

	board.MinX = min(board.MinX, w.Start.X())
	board.MinY = min(board.MinY, w.Start.Y())

	if w.Direction == DirectionHorizontal {
		board.MaxX = max(board.MaxX, w.Start.X()+len(w.Letters))
		board.MaxY = max(board.MaxY, w.Start.Y()+1)
	} else {
		board.MaxY = max(board.MaxY, w.Start.Y()+len(w.Letters))
		board.MaxX = max(board.MaxX, w.Start.X()+1)
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
		_, _ = board.PlaceWord(word)
	}

	return nil
}

func (board *Board) set(point Point, letter rune) {
	board.Grid[point] = letter
}

func (board *Board) wordFormedByNewLetter(letter rune, point Point, direction Direction) (Word, bool) {
	s := string(letter)
	blanks := make(map[Point]struct{})

	start := point

	// starting as far before as possible
	for {
		letter, isSet := board.GetLetter(start.Offset(direction.Vector(-1)))
		if !isSet {
			break
		}

		s = string(letter) + s
		start = start.Offset(direction.Vector(-1))
	}

	// now go as far after as possible
	end := point
	for {
		letter, isSet := board.GetLetter(end.Offset(direction.Vector(1)))
		if !isSet {
			break
		}

		s += string(letter)
		end := end.Offset(direction.Vector(1))
		if _, isBlank := board.Blanks[end]; isBlank {
			blanks[end] = struct{}{}
		}
	}

	if len(s) == 1 {
		return Word{}, false
	}

	newWord := NewWord(start, direction, s)

	for point := range blanks {
		newWord = newWord.WithBlanks(point)
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
	minX := -16
	maxX := 16
	minY := -16
	maxY := 16

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
