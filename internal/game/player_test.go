package game_test

import (
	"testing"

	"github.com/carterjs/words/internal/game"
	"github.com/stretchr/testify/assert"
)

func TestPlayer_HasLetters(t *testing.T) {
	tests := map[string]struct {
		player   game.Player
		letters  []rune
		expected bool
	}{
		"no letters": {
			player: game.Player{
				Letters: nil,
			},
			letters:  []rune{'a'},
			expected: false,
		},
		"single letter": {
			player: game.Player{
				Letters: []rune{'a'},
			},
			letters:  []rune{'a'},
			expected: true,
		},
		"multiple letters": {
			player: game.Player{
				Letters: []rune{'a', 'b', 'c'},
			},
			letters:  []rune{'a', 'b'},
			expected: true,
		},
		"missing letters": {
			player: game.Player{
				Letters: []rune{'a', 'b', 'c'},
			},
			letters:  []rune{'a', 'b', 'c', 'd'},
			expected: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.player.HasLetters(test.letters)
			assert.Equal(t, test.expected, got)
		})
	}
}
