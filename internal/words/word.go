package words

type (
	Word struct {
		Start     Point
		Direction Direction
		Letters   []rune
		Blanks    map[Point]struct{}
	}
)

var BlankLetter = '_'

func NewWord(start Point, direction Direction, value string) Word {
	return Word{
		Start:     start,
		Direction: direction,
		Letters:   []rune(value),
	}
}

func (word Word) WithBlanks(points ...Point) Word {
	if word.Blanks == nil {
		word.Blanks = make(map[Point]struct{})
	}

	for _, point := range points {
		word.Blanks[point] = struct{}{}
	}

	return word
}

func (word Word) Index(i int) (Point, rune, bool) {
	point := word.Start.Offset(word.Direction.Vector(i))

	if i < 0 || i >= len(word.Letters) {
		return point, 0, false
	}

	return point, word.Letters[i], true
}

func (word Word) Get(point Point) (rune, bool) {
	x, y := point.X(), point.Y()
	dx, dy := word.Direction.Vector(1)

	if x < word.Start.X() || y < word.Start.Y() {
		return 0, false
	}

	if x >= word.Start.X() && x < word.Start.X()+dx*len(word.Letters) && y == word.Start.Y() {
		return word.Letters[x-word.Start.X()], true
	}

	if y >= word.Start.Y() && y < word.Start.Y()+dy*len(word.Letters) && x == word.Start.X() {
		return word.Letters[y-word.Start.Y()], true
	}

	return 0, false

}

func (word Word) String() string {
	var s string
	for i, letter := range word.Letters {
		p, _, _ := word.Index(i)
		if _, ok := word.Blanks[p]; ok {
			s += string(BlankLetter)
		} else {
			s += string(letter)
		}
	}

	if len(word.Blanks) > 0 {
		s += " (" + string(word.Letters) + ")"
	}

	return s
}
