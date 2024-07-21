package game_test

import (
	"testing"

	"github.com/carterjs/words/internal/game"
	"github.com/stretchr/testify/assert"
)

func TestNewBoard(t *testing.T) {
	tests := map[string]struct {
		words            []game.Word
		expectedError    error
		expectedIndirect []game.Word
	}{
		"no words": {
			words: nil,
		},
		"single word": {
			words: []game.Word{
				{
					X:         0,
					Y:         0,
					Direction: game.DirectionHorizontal,
					Letters:   []rune("hello"),
				},
			},
			expectedError: nil,
		},
		"basic intersection": {
			words: []game.Word{
				{
					X:         0,
					Y:         0,
					Direction: game.DirectionHorizontal,
					Letters:   []rune("hello"),
				},
				{
					X:         4,
					Y:         -1,
					Direction: game.DirectionVertical,
					Letters:   []rune("world"),
				},
			},
			expectedError: nil,
		},
		"parallel words, horizontal": {
			words: []game.Word{
				{
					X:         0,
					Y:         0,
					Direction: game.DirectionHorizontal,
					Letters:   []rune("hello"),
				},
				{
					X:         4,
					Y:         1,
					Direction: game.DirectionHorizontal,
					Letters:   []rune("world"),
				},
			},
			expectedError: nil,
			expectedIndirect: []game.Word{
				{
					X:         4,
					Y:         0,
					Direction: game.DirectionVertical,
					Letters:   []rune("ow"),
				},
			},
		},
		"parallel words, vertical": {
			words: []game.Word{
				{
					X:         0,
					Y:         0,
					Direction: game.DirectionVertical,
					Letters:   []rune("hello"),
				},
				{
					X:         -1,
					Y:         4,
					Direction: game.DirectionVertical,
					Letters:   []rune("world"),
				},
			},
			expectedError: nil,
			expectedIndirect: []game.Word{
				{
					X:         -1,
					Y:         4,
					Direction: game.DirectionHorizontal,
					Letters:   []rune("wo"),
				},
			},
		},
		"perpendicular word extension": {
			words: []game.Word{
				{
					X:         0,
					Y:         0,
					Direction: game.DirectionVertical,
					Letters:   []rune("hello"),
				},
				{
					X:         0,
					Y:         5,
					Direction: game.DirectionHorizontal,
					Letters:   []rune("slouch"),
				},
			},
			expectedError: nil,
			expectedIndirect: []game.Word{
				{
					X:         0,
					Y:         0,
					Direction: game.DirectionVertical,
					Letters:   []rune("hellos"),
				},
			},
		},
		"word not connected": {
			words: []game.Word{
				{
					X:         0,
					Y:         0,
					Direction: game.DirectionHorizontal,
					Letters:   []rune("hello"),
				},
				{
					X:         0,
					Y:         2,
					Direction: game.DirectionHorizontal,
					Letters:   []rune("world"),
				},
			},
			expectedError: game.ErrWordNotConnected,
		},
		"word conflict": {
			words: []game.Word{
				{
					X:         0,
					Y:         0,
					Direction: game.DirectionHorizontal,
					Letters:   []rune("hello"),
				},
				{
					X:         4,
					Y:         0,
					Direction: game.DirectionVertical,
					Letters:   []rune("world"),
				},
			},
			expectedError: game.WordConflictError{
				X:    4,
				Y:    0,
				Want: 'w',
				Got:  'o',
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			board, err := game.NewBoard(test.words)
			assert.ErrorIs(t, test.expectedError, err)
			if err != nil {
				return
			}
			assert.Equal(t, test.words, board.DirectWords())
			assert.Equal(t, test.expectedIndirect, board.IndirectWords())
			assert.Equal(t, append(test.words, test.expectedIndirect...), board.AllWords())

			t.Logf("Direct: %s Indirect: %s\n%s", board.DirectWords(), board.IndirectWords(), board.String())
		})
	}
}

func TestBoard_PlaceWord(t *testing.T) {
	tests := map[string]struct {
		words             []game.Word
		newWord           game.Word
		expectedPlacement game.PlacementResult
		expectedError     error
	}{
		"no words": {
			words: nil,
			newWord: game.Word{
				X:         0,
				Y:         0,
				Direction: game.DirectionHorizontal,
				Letters:   []rune("hello"),
			},
			expectedPlacement: game.PlacementResult{
				LettersUsed: []rune("hello"),
			},
			expectedError: nil,
		},
		"two word, normal overlap": {
			words: []game.Word{
				{
					X:         0,
					Y:         0,
					Direction: game.DirectionHorizontal,
					Letters:   []rune("hello"),
				},
			},
			newWord: game.Word{
				X:         0,
				Y:         0,
				Direction: game.DirectionVertical,
				Letters:   []rune("hello"),
			},
			expectedPlacement: game.PlacementResult{
				LettersUsed: []rune("ello"),
			},
		},
		"full overlap": {
			words: []game.Word{
				{
					X:         0,
					Y:         0,
					Direction: game.DirectionHorizontal,
					Letters:   []rune("hello"),
				},
			},
			newWord: game.Word{
				X:         0,
				Y:         0,
				Direction: game.DirectionHorizontal,
				Letters:   []rune("hell"),
			},
			expectedPlacement: game.PlacementResult{},
			expectedError:     game.ErrUnchangedBoard,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			board, err := game.NewBoard(test.words)
			assert.NoError(t, err)

			placement, err := board.PlaceWord(test.newWord)
			assert.ErrorIs(t, err, test.expectedError)
			assert.Equal(t, test.expectedPlacement, placement)
		})
	}
}
