package words

import "github.com/google/uuid"

type (
	Word struct {
		ID        string
		X         int
		Y         int
		Direction Direction
		Letters   []rune
		Blanks    map[int]struct{}
	}
)

var BlankLetter = '_'

func NewWord(x, y int, direction Direction, value string) Word {
	return Word{
		ID:        uuid.NewString(),
		X:         x,
		Y:         y,
		Direction: direction,
		Letters:   []rune(value),
	}
}

func (word Word) WithBlank(index int) Word {
	if word.Blanks == nil {
		word.Blanks = make(map[int]struct{})
	}

	word.Blanks[index] = struct{}{}

	return word
}

func (word Word) Index(i int) (int, int, rune, bool) {
	dx, dy := word.Direction.Vector(i)
	x, y := word.X+dx, word.Y+dy

	if i < 0 || i >= len(word.Letters) {
		return x, y, 0, false
	}

	return x, y, word.Letters[i], true
}

func (word Word) Get(x, y int) (rune, bool) {
	dx, dy := word.Direction.Vector(1)

	if x < word.X || y < word.Y {
		return 0, false
	}

	if x >= word.X && x < word.X+dx*len(word.Letters) && y == word.Y {
		return word.Letters[x-word.X], true
	}

	if y >= word.Y && y < word.Y+dy*len(word.Letters) && x == word.X {
		return word.Letters[y-word.Y], true
	}

	return 0, false

}

func (word Word) String() string {
	var s string
	for i, letter := range word.Letters {
		if _, ok := word.Blanks[i]; ok {
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
