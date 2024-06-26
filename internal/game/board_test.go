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
			words: []game.Word{},
		},
		"single word": {
			words: []game.Word{
				{
					X:         0,
					Y:         0,
					Direction: game.DirectionHorizontal,
					Value:     "hello",
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
					Value:     "hello",
				},
				{
					X:         4,
					Y:         -1,
					Direction: game.DirectionVertical,
					Value:     "world",
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
					Value:     "hello",
				},
				{
					X:         4,
					Y:         1,
					Direction: game.DirectionHorizontal,
					Value:     "world",
				},
			},
			expectedError: nil,
			expectedIndirect: []game.Word{
				{
					X:         4,
					Y:         0,
					Direction: game.DirectionVertical,
					Value:     "ow",
				},
			},
		},
		"parallel words, vertical": {
			words: []game.Word{
				{
					X:         0,
					Y:         0,
					Direction: game.DirectionVertical,
					Value:     "hello",
				},
				{
					X:         -1,
					Y:         4,
					Direction: game.DirectionVertical,
					Value:     "world",
				},
			},
			expectedError: nil,
			expectedIndirect: []game.Word{
				{
					X:         -1,
					Y:         4,
					Direction: game.DirectionHorizontal,
					Value:     "wo",
				},
			},
		},
		"perpendicular word extension": {
			words: []game.Word{
				{
					X:         0,
					Y:         0,
					Direction: game.DirectionVertical,
					Value:     "hello",
				},
				{
					X:         0,
					Y:         5,
					Direction: game.DirectionHorizontal,
					Value:     "slouch",
				},
			},
			expectedError: nil,
			expectedIndirect: []game.Word{
				{
					X:         0,
					Y:         0,
					Direction: game.DirectionVertical,
					Value:     "hellos",
				},
			},
		},
		"word not connected": {
			words: []game.Word{
				{
					X:         0,
					Y:         0,
					Direction: game.DirectionHorizontal,
					Value:     "hello",
				},
				{
					X:         0,
					Y:         2,
					Direction: game.DirectionHorizontal,
					Value:     "world",
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
					Value:     "hello",
				},
				{
					X:         4,
					Y:         0,
					Direction: game.DirectionVertical,
					Value:     "world",
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

			// Print for debugging
			t.Logf("Direct: %s Indirect: %s\n%s", board.DirectWords(), board.IndirectWords(), board.String())
		})
	}
}
