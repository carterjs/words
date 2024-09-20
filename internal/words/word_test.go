package words_test

import (
	"github.com/carterjs/words/internal/words"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWord_Index(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		word           words.Word
		index          int
		expectedPoint  words.Point
		expectedLetter rune
		expectedResult bool
	}{
		{
			name:           "horizontal first letter",
			word:           words.NewWord(words.NewPoint(0, 0), words.DirectionHorizontal, "hello"),
			index:          0,
			expectedPoint:  words.NewPoint(0, 0),
			expectedLetter: 'h',
			expectedResult: true,
		},
		{
			name:           "vertical second letter with non-zero start",
			word:           words.NewWord(words.NewPoint(0, 1), words.DirectionVertical, "hello"),
			index:          1,
			expectedPoint:  words.NewPoint(0, 2),
			expectedLetter: 'e',
			expectedResult: true,
		},
		{
			name:           "horizontal last letter",
			word:           words.NewWord(words.NewPoint(0, 0), words.DirectionHorizontal, "hello"),
			index:          4,
			expectedPoint:  words.NewPoint(4, 0),
			expectedLetter: 'o',
			expectedResult: true,
		},
		{
			name:           "out of bounds",
			word:           words.NewWord(words.NewPoint(0, 0), words.DirectionHorizontal, "hello"),
			index:          5,
			expectedPoint:  words.NewPoint(5, 0),
			expectedLetter: 0,
			expectedResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			point, letter, ok := test.word.Index(test.index)
			assert.Equal(t, test.expectedPoint, point)
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
		point          words.Point
		expectedLetter rune
		expectedResult bool
	}{
		{
			name:           "horizontal first letter",
			word:           words.NewWord(words.NewPoint(0, 0), words.DirectionHorizontal, "hello"),
			point:          words.NewPoint(0, 0),
			expectedLetter: 'h',
			expectedResult: true,
		},
		{
			name:           "vertical second letter with non-zero start",
			word:           words.NewWord(words.NewPoint(0, 1), words.DirectionVertical, "hello"),
			point:          words.NewPoint(0, 2),
			expectedLetter: 'e',
			expectedResult: true,
		},
		{
			name:           "horizontal last letter",
			word:           words.NewWord(words.NewPoint(0, 0), words.DirectionHorizontal, "hello"),
			point:          words.NewPoint(4, 0),
			expectedLetter: 'o',
			expectedResult: true,
		},
		{
			name:           "out of bounds",
			word:           words.NewWord(words.NewPoint(0, 0), words.DirectionHorizontal, "hello"),
			point:          words.NewPoint(5, 0),
			expectedLetter: 0,
			expectedResult: false,
		},
		{
			name:           "out of bounds outside",
			word:           words.NewWord(words.NewPoint(0, 0), words.DirectionHorizontal, "hello"),
			point:          words.NewPoint(0, -1),
			expectedLetter: 0,
			expectedResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			letter, ok := test.word.Get(test.point)
			assert.Equal(t, test.expectedLetter, letter)
			assert.Equal(t, test.expectedResult, ok)
		})
	}
}
