package words

// Word is a run of letters laid on the board in a single direction.
type Word struct {
	start     Point
	direction Direction
	letters   []rune
	blanks    map[Point]struct{}
}

// BlankLetter is the rune representing a blank tile in racks and pools.
const BlankLetter = '_'

// NewWord returns a word starting at the given point and running in the
// given direction.
func NewWord(start Point, direction Direction, letters string) Word {
	return Word{
		start:     start,
		direction: direction,
		letters:   []rune(letters),
	}
}

// Start returns the point of the word's first letter.
func (word Word) Start() Point {
	return word.start
}

// Direction returns the direction the word runs in.
func (word Word) Direction() Direction {
	return word.direction
}

// Letters returns the word's letters in order.
func (word Word) Letters() []rune {
	letters := make([]rune, len(word.letters))
	copy(letters, word.letters)
	return letters
}

// Length returns the number of letters in the word.
func (word Word) Length() int {
	return len(word.letters)
}

// Blanks returns the points of the word occupied by blank tiles.
func (word Word) Blanks() []Point {
	points := make([]Point, 0, len(word.blanks))
	for position := range word.Length() {
		point, _, _ := word.Index(position)
		if _, isBlank := word.blanks[point]; isBlank {
			points = append(points, point)
		}
	}
	return points
}

// Blank reports whether the given point of the word holds a blank tile.
func (word Word) Blank(point Point) bool {
	_, isBlank := word.blanks[point]
	return isBlank
}

// WithBlanks returns a copy of the word with the given points marked as
// blank tiles.
func (word Word) WithBlanks(points ...Point) Word {
	blanks := make(map[Point]struct{}, len(word.blanks)+len(points))
	for point := range word.blanks {
		blanks[point] = struct{}{}
	}
	for _, point := range points {
		blanks[point] = struct{}{}
	}

	word.blanks = blanks
	return word
}

// Index returns the point and letter at the given position, and whether the
// position is within the word.
func (word Word) Index(position int) (Point, rune, bool) {
	point := word.start.Offset(word.direction.Vector(position))

	if position < 0 || position >= len(word.letters) {
		return point, 0, false
	}

	return point, word.letters[position], true
}

// At returns the letter at the given point and whether the word covers it.
func (word Word) At(point Point) (rune, bool) {
	for position := range word.letters {
		wordPoint, letter, _ := word.Index(position)
		if wordPoint == point {
			return letter, true
		}
	}

	return 0, false
}

// String renders the word's letters, masking blanks and appending the
// underlying letters when any blank is present.
func (word Word) String() string {
	var rendered string
	for position, letter := range word.letters {
		point, _, _ := word.Index(position)
		if _, isBlank := word.blanks[point]; isBlank {
			rendered += string(BlankLetter)
		} else {
			rendered += string(letter)
		}
	}

	if len(word.blanks) > 0 {
		rendered += " (" + string(word.letters) + ")"
	}

	return rendered
}
