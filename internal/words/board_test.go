package words_test

import (
	"github.com/carterjs/words/internal/pattern"
	"github.com/carterjs/words/internal/words"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoard_PlaceWord(t *testing.T) {
	tests := map[string]struct {
		words             []words.Word
		modifiers         pattern.Group[words.Modifier]
		newWord           words.Word
		expectedPlacement words.PlacementResult
		expectedError     error
	}{
		"no words": {
			words:   nil,
			newWord: words.NewWord(0, 0, words.DirectionHorizontal, "hello"),
			expectedPlacement: words.PlacementResult{
				DirectWord:  words.NewWord(0, 0, words.DirectionHorizontal, "hello"),
				LettersUsed: []rune("hello"),
			},
			expectedError: nil,
		},
		"two word, normal overlap": {
			words: []words.Word{
				words.NewWord(0, 0, words.DirectionHorizontal, "hello"),
			},
			newWord: words.Word{
				X:         0,
				Y:         0,
				Direction: words.DirectionVertical,
				Letters:   []rune("hello"),
			},
			expectedPlacement: words.PlacementResult{
				DirectWord:  words.NewWord(0, 0, words.DirectionVertical, "hello"),
				LettersUsed: []rune("ello"),
			},
		},
		"full overlap": {
			words: []words.Word{
				{
					X:         0,
					Y:         0,
					Direction: words.DirectionHorizontal,
					Letters:   []rune("hello"),
				},
			},
			newWord: words.Word{
				X:         0,
				Y:         0,
				Direction: words.DirectionHorizontal,
				Letters:   []rune("hell"),
			},
			expectedPlacement: words.PlacementResult{},
			expectedError:     words.ErrUnchanged,
		},
		"doubled word": {
			words: []words.Word{
				words.NewWord(0, 0, words.DirectionHorizontal, "hello"),
			},
			modifiers: pattern.Group[words.Modifier]{
				{
					Value: words.ModifierDoubleWord,
					Explicit: []pattern.Explicit{
						{
							X: 1,
							Y: 0,
						},
					},
				},
			},
			newWord: words.NewWord(0, 0, words.DirectionVertical, "hello"),
			expectedPlacement: words.PlacementResult{
				LettersUsed: []rune("ello"),
				DirectWord:  words.NewWord(0, 0, words.DirectionVertical, "hello"),
				Modifiers: map[int]words.Modifier{
					1: words.ModifierDoubleWord,
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			testBoard, err := words.newBoard("blah", test.modifiers)
			assert.NoError(t, err)

			for _, word := range test.words {
				_, err := testBoard.placeWord(word)
				assert.NoError(t, err)
			}

			placement, err := testBoard.placeWord(test.newWord)
			assert.ErrorIs(t, err, test.expectedError)
			assert.Equal(t, test.expectedPlacement, placement)
		})
	}
}

type MockModifierGetter struct {
	GetFunc func(x, y int) (words.Modifier, bool)
}

func newEmptyModifierGetter() MockModifierGetter {
	return MockModifierGetter{
		GetFunc: func(x, y int) (words.Modifier, bool) {
			return "", false
		},
	}
}

func (m MockModifierGetter) Get(x, y int) (words.Modifier, bool) {
	return m.GetFunc(x, y)
}
