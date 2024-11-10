package words_test

import (
	"github.com/carterjs/words/internal/words"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGame_FindPlacements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		game          *words.Game
		playerID      string
		point         words.Point
		word          string
		expected      []words.PlacementResult
		expectedError string
	}{
		{
			name: "first word",
			game: newStartedGame(t, []words.Player{
				{
					ID:      "player1",
					Letters: []rune("HI"),
				},
			}),
			playerID: "player1",
			point:    words.NewPoint(0, 0),
			word:     "HI",
			expected: []words.PlacementResult{
				{
					LettersUsed: map[words.Point]rune{
						words.NewPoint(0, 0): 'H',
						words.NewPoint(1, 0): 'I',
					},
					DirectWord: words.NewWord(words.NewPoint(0, 0), words.DirectionHorizontal, "HI"),
					Points:     5,
				},
				{
					LettersUsed: map[words.Point]rune{
						words.NewPoint(-1, 0): 'H',
						words.NewPoint(0, 0):  'I',
					},
					DirectWord: words.NewWord(words.NewPoint(-1, 0), words.DirectionHorizontal, "HI"),
					Points:     5,
				},
				{
					LettersUsed: map[words.Point]rune{
						words.NewPoint(0, 0): 'H',
						words.NewPoint(0, 1): 'I',
					},
					DirectWord: words.NewWord(words.NewPoint(0, 0), words.DirectionVertical, "HI"),
					Points:     5,
				},
				{
					LettersUsed: map[words.Point]rune{
						words.NewPoint(0, -1): 'H',
						words.NewPoint(0, 0):  'I',
					},
					DirectWord: words.NewWord(words.NewPoint(0, -1), words.DirectionVertical, "HI"),
					Points:     5,
				},
			},
		},
		{
			name: "player can't play word",
			game: newStartedGame(t, []words.Player{
				{
					ID:      "player1",
					Letters: []rune("HELLO1234567"),
				},
			},
				words.NewWord(words.NewPoint(0, 0), words.DirectionHorizontal, "HELLO"),
			),
			playerID:      "player1",
			point:         words.NewPoint(0, 0),
			word:          "WORLD",
			expectedError: words.ErrCannotPlayWord.Error(),
		},
		{
			name: "multiple options sorted by points",
			game: newStartedGame(t, []words.Player{
				{
					ID:      "player1",
					Letters: []rune("YOUTOO4567"),
				},
			},
				words.NewWord(words.NewPoint(0, 0), words.DirectionHorizontal, "YOU"),
			),
			playerID: "player1",
			point:    words.NewPoint(1, 0),
			word:     "TOO",
			expected: []words.PlacementResult{
				{
					LettersUsed: map[words.Point]rune{
						words.NewPoint(1, -1): 'T',
						words.NewPoint(1, 1):  'O',
					},
					DirectWord: words.NewWord(words.NewPoint(1, -1), words.DirectionVertical, "TOO"),
					Points:     5,
					Modifiers: map[int]words.Modifier{
						0: words.ModifierDoubleLetter,
						2: words.ModifierDoubleLetter,
					},
				},
				{
					LettersUsed: map[words.Point]rune{
						words.NewPoint(1, -2): 'T',
						words.NewPoint(1, -1): 'O',
					},
					DirectWord: words.NewWord(words.NewPoint(1, -2), words.DirectionVertical, "TOO"),
					Points:     4,
					Modifiers: map[int]words.Modifier{
						1: words.ModifierDoubleLetter,
					},
				},
			},
		},
		{
			name: "blanks",
			game: newStartedGame(t, []words.Player{
				{
					ID:      "player1",
					Letters: []rune("TEST__T4567"),
				},
			},
				words.NewWord(words.NewPoint(0, 0), words.DirectionHorizontal, "TEST"),
			),
			playerID: "player1",
			point:    words.NewPoint(0, 0),
			word:     "TEST",
			expected: []words.PlacementResult{
				{
					LettersUsed: map[words.Point]rune{
						words.NewPoint(0, 1): '_',
						words.NewPoint(0, 2): '_',
						words.NewPoint(0, 3): 'T',
					},
					DirectWord: words.NewWord(words.NewPoint(0, 0), words.DirectionVertical, "TEST").WithBlanks(
						words.NewPoint(0, 1),
						words.NewPoint(0, 2),
					),
					Points: 2,
				},
				{
					LettersUsed: map[words.Point]rune{
						words.NewPoint(0, -3): 'T',
						words.NewPoint(0, -2): '_',
						words.NewPoint(0, -1): '_',
					},
					DirectWord: words.NewWord(words.NewPoint(0, -3), words.DirectionVertical, "TEST").WithBlanks(
						words.NewPoint(0, -2),
						words.NewPoint(0, -1),
					),
					Points: 2,
				},
			},
		},
		{
			name: "all blanks",
			game: newStartedGame(t, []words.Player{
				{
					ID:      "player1",
					Letters: []rune("_______"),
				},
			}),
			playerID: "player1",
			point:    words.NewPoint(-1, 0),
			word:     "BRO",
			expected: []words.PlacementResult{
				{
					LettersUsed: map[words.Point]rune{
						words.NewPoint(-1, 0): '_',
						words.NewPoint(0, 0):  '_',
						words.NewPoint(1, 0):  '_',
					},
					DirectWord: words.NewWord(words.NewPoint(-1, 0), words.DirectionHorizontal, "BRO").WithBlanks(
						words.NewPoint(-1, 0),
						words.NewPoint(0, 0),
						words.NewPoint(1, 0),
					),
				},
				{
					LettersUsed: map[words.Point]rune{
						words.NewPoint(-2, 0): '_',
						words.NewPoint(-1, 0): '_',
						words.NewPoint(0, 0):  '_',
					},
					DirectWord: words.NewWord(words.NewPoint(-2, 0), words.DirectionHorizontal, "BRO").WithBlanks(
						words.NewPoint(-2, 0),
						words.NewPoint(-1, 0),
						words.NewPoint(0, 0),
					),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			placements, err := test.game.FindPlacements(test.playerID, test.point, test.word)
			if test.expectedError != "" {
				assert.EqualError(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, placements)
			}
		})
	}
}

func newStartedGame(t *testing.T, players []words.Player, w ...words.Word) *words.Game {
	game := words.NewGame(words.Presets[0].Config)
	game.Players = players

	err := game.Start()
	assert.NoError(t, err)

	for i, word := range w {
		_, err := game.PlayWord(game.Players[i%len(players)].ID, word)
		assert.NoError(t, err)
	}

	return game
}
