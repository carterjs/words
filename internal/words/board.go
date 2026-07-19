package words

import (
	"fmt"
	"strings"
)

// Board holds the letters placed so far on the unbounded grid.
type Board struct {
	grid   map[Point]rune
	blanks map[Point]struct{}
	words  []Word
	bounds Bounds
	config Config
}

// Bounds is the smallest rectangle containing every placed letter, expressed
// as an inclusive minimum and exclusive maximum on each axis.
type Bounds struct {
	MinX int `json:"minX"`
	MinY int `json:"minY"`
	MaxX int `json:"maxX"`
	MaxY int `json:"maxY"`
}

// NewBoard returns an empty board using the given configuration's modifiers.
func NewBoard(config Config) *Board {
	return &Board{
		grid:   make(map[Point]rune),
		blanks: make(map[Point]struct{}),
		config: config,
	}
}

// Letter returns the letter at the given point and whether one is placed there.
func (board *Board) Letter(point Point) (rune, bool) {
	letter, exists := board.grid[point]
	return letter, exists
}

// PlaceholderLetter marks a position in a typed word that must be filled by
// a letter already on the board.
const PlaceholderLetter = '*'

// FillPlaceholders replaces each placeholder in the word with the board
// letter at its position, reporting false when one lands on an empty cell.
func (board *Board) FillPlaceholders(word Word) (Word, bool) {
	letters := word.Letters()

	changed := false
	for position, letter := range letters {
		if letter != PlaceholderLetter {
			continue
		}

		point, _, _ := word.Index(position)
		existing, occupied := board.Letter(point)
		if !occupied {
			return Word{}, false
		}

		letters[position] = existing
		changed = true
	}

	if !changed {
		return word, true
	}

	return NewWord(word.Start(), word.Direction(), string(letters)), true
}

// Modifier returns the modifier at the given point and whether one exists there.
func (board *Board) Modifier(point Point) (Modifier, bool) {
	return board.config.Modifiers.Get(point.Column(), point.Row())
}

// Bounds returns the smallest rectangle containing every placed letter.
func (board *Board) Bounds() Bounds {
	return board.bounds
}

// Words returns the words placed on the board in play order.
func (board *Board) Words() []Word {
	placed := make([]Word, len(board.words))
	copy(placed, board.words)
	return placed
}

// PlaceWord validates the word against the board and places it, returning
// the letters used, indirect words formed, and points scored.
func (board *Board) PlaceWord(word Word) (PlacementResult, error) {
	result, err := board.tryWordPlacement(word)
	if err != nil {
		return PlacementResult{}, fmt.Errorf("checking placement: %w", err)
	}

	// place the fully-resolved word so blank markings survive
	word = result.DirectWord

	for position := range word.Length() {
		point, letter, _ := word.Index(position)
		board.grid[point] = letter
		if word.Blank(point) {
			board.blanks[point] = struct{}{}
		}
	}

	board.expandBounds(word)
	board.words = append(board.words, word)

	return result, nil
}

func (board *Board) tryWordPlacement(word Word) (PlacementResult, error) {
	needsConnection := len(board.grid) > 0
	if !needsConnection {
		if _, intersectsCenter := word.At(NewPoint(0, 0)); !intersectsCenter {
			return PlacementResult{}, ErrFirstWordNotCentered
		}
	}

	result := PlacementResult{
		LettersUsed: make(map[Point]rune),
		DirectWord:  word,
	}

	connected, err := board.applyWordLetters(word, &result)
	if err != nil {
		return PlacementResult{}, fmt.Errorf("checking letters: %w", err)
	}

	if needsConnection && !connected {
		return PlacementResult{}, ErrWordNotConnected
	}

	if len(result.LettersUsed) == 0 {
		return PlacementResult{}, ErrUnchanged
	}

	if err := board.assertWordBoundaries(word); err != nil {
		return PlacementResult{}, fmt.Errorf("checking boundaries: %w", err)
	}

	return placementWithPoints(result, board.config.LetterPoints), nil
}

// applyWordLetters walks the word's cells, recording spent letters, modifiers
// hit, and indirect words into result. It reports whether the word touches
// any letter already on the board.
func (board *Board) applyWordLetters(word Word, result *PlacementResult) (bool, error) {
	var connected bool

	for position := range word.Length() {
		point, letter, _ := word.Index(position)

		if currentLetter, occupied := board.Letter(point); occupied {
			if currentLetter != letter {
				return false, WordConflictError{
					column: point.Column(),
					row:    point.Row(),
					want:   letter,
					got:    currentLetter,
				}
			}

			// overlapping a blank makes this word's letter a blank for scoring
			if _, isBlank := board.blanks[point]; isBlank {
				result.DirectWord = result.DirectWord.WithBlanks(point)
			}

			connected = true
			continue
		}

		if modifier, hasModifier := board.Modifier(point); hasModifier {
			if result.Modifiers == nil {
				result.Modifiers = make(map[int]Modifier)
			}
			result.Modifiers[position] = modifier
		}

		if word.Blank(point) {
			result.LettersUsed[point] = BlankLetter
		} else {
			result.LettersUsed[point] = letter
		}

		if indirectWord, hasIndirectWord := board.wordFormedByNewLetter(letter, point, word.Direction().Other()); hasIndirectWord {
			result.IndirectWords = append(result.IndirectWords, indirectWord)
			connected = true
		}
	}

	return connected, nil
}

// assertWordBoundaries rejects placements that run into letters immediately
// before or after the word, which would silently form a longer word.
func (board *Board) assertWordBoundaries(word Word) error {
	deltaColumn, deltaRow := word.Direction().Vector(1)

	before := word.Start().Offset(-deltaColumn, -deltaRow)
	if _, occupied := board.Letter(before); occupied {
		return ErrIncomplete
	}

	after := word.Start().Offset(word.Direction().Vector(word.Length()))
	if _, occupied := board.Letter(after); occupied {
		return ErrIncomplete
	}

	return nil
}

func (board *Board) expandBounds(word Word) {
	start := word.Start()
	board.bounds.MinX = min(board.bounds.MinX, start.Column())
	board.bounds.MinY = min(board.bounds.MinY, start.Row())

	if word.Direction() == DirectionHorizontal {
		board.bounds.MaxX = max(board.bounds.MaxX, start.Column()+word.Length())
		board.bounds.MaxY = max(board.bounds.MaxY, start.Row()+1)
	} else {
		board.bounds.MaxY = max(board.bounds.MaxY, start.Row()+word.Length())
		board.bounds.MaxX = max(board.bounds.MaxX, start.Column()+1)
	}
}

// removeLastWord takes the most recently placed word off the board by
// replaying the remaining words onto a fresh grid.
func (board *Board) removeLastWord() error {
	if len(board.words) == 0 {
		return ErrUnchanged
	}

	remaining := board.words[:len(board.words)-1]

	board.grid = make(map[Point]rune)
	board.blanks = make(map[Point]struct{})
	board.words = nil
	board.bounds = Bounds{}

	for _, word := range remaining {
		if _, err := board.PlaceWord(word); err != nil {
			return fmt.Errorf("replaying word %q: %w", word.String(), err)
		}
	}

	return nil
}

// wordFormedByNewLetter finds the perpendicular word completed by placing
// the given letter, scanning for existing letters on both sides of it.
func (board *Board) wordFormedByNewLetter(letter rune, point Point, direction Direction) (Word, bool) {
	start := point
	for {
		previous := start.Offset(direction.Vector(-1))
		if _, occupied := board.Letter(previous); !occupied {
			break
		}
		start = previous
	}

	end := point
	for {
		next := end.Offset(direction.Vector(1))
		if _, occupied := board.Letter(next); !occupied {
			break
		}
		end = next
	}

	if start == end {
		return Word{}, false
	}

	var letters []rune
	var blankPoints []Point
	for current := start; ; current = current.Offset(direction.Vector(1)) {
		if current == point {
			letters = append(letters, letter)
		} else {
			currentLetter, _ := board.Letter(current)
			letters = append(letters, currentLetter)
			if _, isBlank := board.blanks[current]; isBlank {
				blankPoints = append(blankPoints, current)
			}
		}

		if current == end {
			break
		}
	}

	return NewWord(start, direction, string(letters)).WithBlanks(blankPoints...), true
}

const (
	renderCellWidth  = 3
	renderMinExtent  = 16
	renderEdgeBuffer = 1
)

// String renders the board as a text grid for debugging.
func (board *Board) String() string {
	minX, minY, maxX, maxY := board.renderExtents()

	var builder strings.Builder

	writeColumnHeading(&builder, minX, maxX)

	for row := minY; row <= maxY; row++ {
		builder.WriteString(centered(fmt.Sprint(row), renderCellWidth))
		builder.WriteRune('│')
		for column := minX; column <= maxX; column++ {
			board.writeCell(&builder, column, row)
			builder.WriteRune('│')
		}
		builder.WriteString(centered(fmt.Sprint(row), renderCellWidth))
		builder.WriteRune('\n')
	}

	writeColumnHeading(&builder, minX, maxX)

	return builder.String()
}

func (board *Board) renderExtents() (int, int, int, int) {
	minX, minY := -renderMinExtent, -renderMinExtent
	maxX, maxY := renderMinExtent, renderMinExtent

	if board.bounds.MinX <= minX {
		minX = board.bounds.MinX - renderEdgeBuffer
	}
	if board.bounds.MaxX >= maxX {
		maxX = board.bounds.MaxX - 1 + renderEdgeBuffer
	}
	if board.bounds.MinY <= minY {
		minY = board.bounds.MinY - renderEdgeBuffer
	}
	if board.bounds.MaxY >= maxY {
		maxY = board.bounds.MaxY - 1 + renderEdgeBuffer
	}

	return minX, minY, maxX, maxY
}

func (board *Board) writeCell(builder *strings.Builder, column, row int) {
	if letter, occupied := board.Letter(NewPoint(column, row)); occupied {
		builder.WriteString(centered(string(letter), renderCellWidth))
		return
	}

	if modifier, hasModifier := board.config.Modifiers.Get(column, row); hasModifier {
		builder.WriteString(centered(string(modifier), renderCellWidth))
	} else if column == 0 && row == 0 {
		builder.WriteString(strings.Repeat("▒", renderCellWidth))
	} else {
		builder.WriteString(strings.Repeat("░", renderCellWidth))
	}
}

func writeColumnHeading(builder *strings.Builder, minX, maxX int) {
	builder.WriteString(centered(" ", renderCellWidth+1))
	for column := minX; column <= maxX; column++ {
		builder.WriteString(centered(fmt.Sprint(column), renderCellWidth))
		builder.WriteRune(' ')
	}
	builder.WriteRune('\n')
}

func centered(text string, width int) string {
	return fmt.Sprintf("%*s", -width, fmt.Sprintf("%*s", (width+len(text))/2, text))
}
