package game_test

import (
	"testing"

	"github.com/carterjs/words/internal/game"
	"github.com/stretchr/testify/assert"
)

func TestGame_PassphraseMatches(t *testing.T) {
	tests := map[string]struct {
		game       game.Game
		passphrase string
		expected   bool
	}{
		"match": {
			game: game.Game{
				PassphraseHash: game.HashPassphrase("password"),
			},
			passphrase: "password",
			expected:   true,
		},
		"no match": {
			game: game.Game{
				PassphraseHash: game.HashPassphrase("password"),
			},
			passphrase: "notpassword",
			expected:   false,
		},
		"empty": {
			game: game.Game{
				PassphraseHash: game.HashPassphrase(""),
			},
			passphrase: "",
			expected:   true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.game.PassphraseMatches(test.passphrase) != test.expected {
				t.Errorf("expected %t", test.expected)
			}
		})
	}

}

func TestGame_TakeLetters(t *testing.T) {
	tests := map[string]struct {
		game              game.Game
		n                 int
		expectedLetters   []rune
		expectedError     error
		expectedRemaining []rune
	}{
		"no letters": {
			game: game.Game{
				Letters: nil,
			},
			n:                 1,
			expectedError:     game.ErrNoLettersInPool,
			expectedRemaining: nil,
		},
		"take all": {
			game: game.Game{
				Letters: []rune("abc"),
			},
			n:                 3,
			expectedLetters:   []rune("abc"),
			expectedRemaining: []rune{},
		},
		"take some": {
			game: game.Game{
				Letters: []rune("abc"),
			},
			n:                 2,
			expectedLetters:   []rune("ab"),
			expectedRemaining: []rune("c"),
		},
		"take more than available": {
			game: game.Game{
				Letters: []rune("abc"),
			},
			n:                 4,
			expectedLetters:   []rune("abc"),
			expectedRemaining: []rune{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			letters, err := (&test.game).TakeLetters(test.n)
			assert.ErrorIs(t, err, test.expectedError)
			assert.Equal(t, test.expectedLetters, letters)
			assert.Equal(t, test.expectedRemaining, test.game.Letters)
		})
	}
}
