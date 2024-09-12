package words_test

import (
	"github.com/carterjs/words/internal/words"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlacementResult_GetScore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		placementResult words.PlacementResult
		letterPoints    map[rune]int
		expectedScore   int
	}{
		{
			name: "basic",
			placementResult: words.PlacementResult{
				DirectWord: words.NewWord(0, 0, words.DirectionHorizontal, "hello"),
			},
			letterPoints: map[rune]int{
				'h': 1,
				'e': 1,
				'l': 1,
				'o': 1,
			},
			expectedScore: 5,
		},
		{
			name: "indirects",
			placementResult: words.PlacementResult{
				DirectWord: words.NewWord(0, 0, words.DirectionHorizontal, "hello"),
				IndirectWords: []words.Word{
					words.NewWord(0, 0, words.DirectionVertical, "world"),
				},
			},
			letterPoints: map[rune]int{
				'h': 1,
				'e': 1,
				'l': 1,
				'o': 1,
				'w': 2,
				'r': 2,
				'd': 2,
			},
			expectedScore: 13,
		},
		{
			name: "blanks and Modifiers",
			placementResult: words.PlacementResult{
				DirectWord: words.NewWord(0, 0, words.DirectionHorizontal, "hello").WithBlank(0),
				IndirectWords: []words.Word{
					words.NewWord(0, 0, words.DirectionVertical, "world").WithBlank(0),
				},
				Modifiers: map[int]words.Modifier{
					1: words.ModifierTripleWord,
				},
			},
			letterPoints: map[rune]int{
				'h': 10,
				'e': 1,
				'l': 1,
				'o': 1,
				'w': 1,
				'r': 1,
				'd': 1,
			},
			expectedScore: 16,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			score := test.placementResult.Points
			assert.Equal(t, test.expectedScore, score)
		})
	}
}
