package words_test

import (
	"testing"

	"github.com/carterjs/words/internal/words"
	"github.com/stretchr/testify/assert"
)

func TestWord_At(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		word       words.Word
		point      words.Point
		wantLetter rune
		wantFound  bool
	}{
		{name: "finds a letter in a horizontal word", word: horizontal(0, 0, "ABC"), point: words.NewPoint(1, 0), wantLetter: 'B', wantFound: true},
		{name: "finds a letter in a vertical word", word: vertical(2, -1, "ABC"), point: words.NewPoint(2, 1), wantLetter: 'C', wantFound: true},
		{name: "misses a point off the word's line", word: horizontal(0, 0, "ABC"), point: words.NewPoint(1, 1)},
		{name: "misses a point past the word's end", word: horizontal(0, 0, "ABC"), point: words.NewPoint(3, 0)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			letter, found := test.word.At(test.point)

			assert.Equal(t, test.wantFound, found)
			assert.Equal(t, test.wantLetter, letter)
		})
	}
}

func TestWord_WithBlanks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		blanks     []words.Point
		wantBlanks int
		wantString string
	}{
		{name: "keeps a word without blanks intact", wantString: "ABC"},
		{name: "masks blanks and appends the letters", blanks: []words.Point{words.NewPoint(1, 0)}, wantBlanks: 1, wantString: "A_C (ABC)"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			original := horizontal(0, 0, "ABC")
			blanked := original.WithBlanks(test.blanks...)

			assert.Len(t, blanked.Blanks(), test.wantBlanks)
			assert.Equal(t, test.wantString, blanked.String())
			// the original word is never mutated
			assert.Empty(t, original.Blanks())
		})
	}
}
