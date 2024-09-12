package words_test

import (
	"github.com/carterjs/words/internal/words"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWord_Index(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		word                 words.Word
		index                int
		expectedX, expectedY int
		expectedLetter       rune
		expectedResult       bool
	}{
		{
			name:           "horizontal first letter",
			word:           words.NewWord(0, 0, words.DirectionHorizontal, "hello"),
			index:          0,
			expectedX:      0,
			expectedY:      0,
			expectedLetter: 'h',
			expectedResult: true,
		},
		{
			name:           "vertical second letter with non-zero start",
			word:           words.NewWord(0, 1, words.DirectionVertical, "hello"),
			index:          1,
			expectedX:      0,
			expectedY:      2,
			expectedLetter: 'e',
			expectedResult: true,
		},
		{
			name:           "horizontal last letter",
			word:           words.NewWord(0, 0, words.DirectionHorizontal, "hello"),
			index:          4,
			expectedX:      4,
			expectedY:      0,
			expectedLetter: 'o',
			expectedResult: true,
		},
		{
			name:           "out of bounds",
			word:           words.NewWord(0, 0, words.DirectionHorizontal, "hello"),
			index:          5,
			expectedX:      5,
			expectedY:      0,
			expectedLetter: 0,
			expectedResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			x, y, letter, ok := test.word.Index(test.index)
			assert.Equal(t, test.expectedX, x)
			assert.Equal(t, test.expectedY, y)
			assert.Equal(t, test.expectedLetter, letter)
			assert.Equal(t, test.expectedResult, ok)
		})
	}
}

func TestWord_Get(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		word           words.Word
		x, y           int
		expectedLetter rune
		expectedResult bool
	}{
		{
			name:           "horizontal first letter",
			word:           words.NewWord(0, 0, words.DirectionHorizontal, "hello"),
			x:              0,
			y:              0,
			expectedLetter: 'h',
			expectedResult: true,
		},
		{
			name:           "vertical second letter with non-zero start",
			word:           words.NewWord(0, 1, words.DirectionVertical, "hello"),
			x:              0,
			y:              2,
			expectedLetter: 'e',
			expectedResult: true,
		},
		{
			name:           "horizontal last letter",
			word:           words.NewWord(0, 0, words.DirectionHorizontal, "hello"),
			x:              4,
			y:              0,
			expectedLetter: 'o',
			expectedResult: true,
		},
		{
			name:           "out of bounds",
			word:           words.NewWord(0, 0, words.DirectionHorizontal, "hello"),
			x:              5,
			y:              0,
			expectedLetter: 0,
			expectedResult: false,
		},
		{
			name:           "out of bounds outside",
			word:           words.NewWord(0, 0, words.DirectionHorizontal, "hello"),
			x:              0,
			y:              -1,
			expectedLetter: 0,
			expectedResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			letter, ok := test.word.Get(test.x, test.y)
			assert.Equal(t, test.expectedLetter, letter)
			assert.Equal(t, test.expectedResult, ok)
		})
	}
}
