package words_test

import (
	"github.com/carterjs/words/internal/pattern"
	"github.com/carterjs/words/internal/words"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBoard_PlaceWord(t *testing.T) {
	tests := map[string]struct {
		words             []words.Word
		config            words.Config
		newWord           words.Word
		expectedPlacement words.PlacementResult
		expectedError     error
	}{
		"one word": {
			words:   nil,
			newWord: words.NewWord(words.NewPoint(0, 0), words.DirectionHorizontal, "HELLO"),
			expectedPlacement: words.PlacementResult{
				DirectWord: words.NewWord(words.NewPoint(0, 0), words.DirectionHorizontal, "HELLO"),
				LettersUsed: map[words.Point]rune{
					words.NewPoint(0, 0): 'H',
					words.NewPoint(1, 0): 'E',
					words.NewPoint(2, 0): 'L',
					words.NewPoint(3, 0): 'L',
					words.NewPoint(4, 0): 'O',
				},
				Points: 5,
			},
			config: words.Config{
				LetterPoints: map[rune]int{
					'H': 1,
					'E': 1,
					'L': 1,
					'O': 1,
				},
			},
			expectedError: nil,
		},
		"two word, normal overlap": {
			words: []words.Word{
				words.NewWord(words.NewPoint(0, 0), words.DirectionHorizontal, "HELLO"),
			},
			newWord: words.NewWord(words.NewPoint(0, 0), words.DirectionVertical, "HELLO"),
			config: words.Config{
				LetterPoints: map[rune]int{
					'H': 1,
					'E': 1,
					'L': 1,
					'O': 1,
				},
			},
			expectedPlacement: words.PlacementResult{
				DirectWord: words.NewWord(words.NewPoint(0, 0), words.DirectionVertical, "HELLO"),
				LettersUsed: map[words.Point]rune{
					"0,1": 'E',
					"0,2": 'L',
					"0,3": 'L',
					"0,4": 'O',
				},
				Points: 5,
			},
		},
		"full overlap": {
			words: []words.Word{
				words.NewWord(words.NewPoint(0, 0), words.DirectionHorizontal, "HELLO"),
			},
			newWord:           words.NewWord(words.NewPoint(0, 0), words.DirectionHorizontal, "HELLO"),
			expectedPlacement: words.PlacementResult{},
			expectedError:     words.ErrUnchanged,
		},
		"doubled word": {
			words: []words.Word{
				words.NewWord(words.NewPoint(0, 0), words.DirectionHorizontal, "HELLO"),
			},
			config: words.Config{
				LetterPoints: map[rune]int{
					'H': 1,
					'E': 1,
					'L': 1,
					'O': 1,
				},
				Modifiers: pattern.Group[words.Modifier]{
					{
						Value: words.ModifierDoubleWord,
						Explicit: []pattern.Explicit{
							{
								X: 0,
								Y: 1,
							},
						},
					},
				},
			},
			newWord: words.NewWord(words.NewPoint(0, 0), words.DirectionVertical, "HELLO"),
			expectedPlacement: words.PlacementResult{
				LettersUsed: map[words.Point]rune{
					words.NewPoint(0, 1): 'E',
					words.NewPoint(0, 2): 'L',
					words.NewPoint(0, 3): 'L',
					words.NewPoint(0, 4): 'O',
				},
				DirectWord: words.NewWord(words.NewPoint(0, 0), words.DirectionVertical, "HELLO"),
				Modifiers: map[int]words.Modifier{
					1: words.ModifierDoubleWord,
				},
				Points: 10,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			testBoard := words.NewBoard("blah", test.config)

			for _, word := range test.words {
				_, err := testBoard.PlaceWord(word)
				assert.NoError(t, err)
			}

			placement, err := testBoard.PlaceWord(test.newWord)
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
