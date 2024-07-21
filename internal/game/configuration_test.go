package game_test

import (
	"sort"
	"testing"

	"github.com/carterjs/words/internal/game"
	"github.com/stretchr/testify/assert"
)

func TestConfiguration_GetLetters(t *testing.T) {
	tests := map[string]struct {
		configuration game.Configuration
		expected      []rune
	}{
		"no letters": {
			configuration: game.Configuration{
				Letters: nil,
			},
			expected: nil,
		},
		"single letter": {
			configuration: game.Configuration{
				Letters: []game.LetterConfiguration{
					{
						Letter: 'a',
						Count:  1,
					},
				},
			},
			expected: []rune{'a'},
		},
		"multiple letters": {
			configuration: game.Configuration{
				Letters: []game.LetterConfiguration{
					{
						Letter: 'a',
						Count:  1,
					},
					{
						Letter: 'b',
						Count:  2,
					},
					{
						Letter: 'c',
						Count:  3,
					},
				},
			},
			expected: []rune{'a', 'b', 'b', 'c', 'c', 'c'},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			letters := test.configuration.GetLetters()
			if len(letters) != len(test.expected) {
				t.Fatalf("expected %d letters, got %d", len(test.expected), len(letters))
			}

			sort.Slice(letters, func(i, j int) bool {
				return letters[i] < letters[j]
			})

			for i, expectedLetter := range test.expected {
				if letters[i] != expectedLetter {
					t.Fatalf("expected letter %q at index %d, got %q", expectedLetter, i, letters[i])
				}
			}
		})
	}
}

func TestConfiguration_GetScoreForWord(t *testing.T) {
	tests := map[string]struct {
		configuration game.Configuration
		word          game.Word
		expected      int
	}{
		"empty word": {
			configuration: game.Configuration{
				Letters: []game.LetterConfiguration{
					{
						Letter: 'a',
						Points: 1,
					},
					{
						Letter: 'b',
						Points: 2,
					},
					{
						Letter: 'c',
						Points: 3,
					},
				},
			},
			word:     game.Word{},
			expected: 0,
		},
		"word with letters not in configuration": {
			configuration: game.Configuration{
				Letters: []game.LetterConfiguration{
					{
						Letter: 'a',
						Points: 1,
					},
					{
						Letter: 'b',
						Points: 2,
					},
					{
						Letter: 'c',
						Points: 3,
					},
				},
			},
			word:     game.Word{Letters: []rune("def")},
			expected: 0,
		},
		"word with letters in configuration": {
			configuration: game.Configuration{
				Letters: []game.LetterConfiguration{
					{
						Letter: 'a',
						Points: 1,
					},
					{
						Letter: 'b',
						Points: 2,
					},
					{
						Letter: 'c',
						Points: 3,
					},
				},
			},
			word:     game.Word{Letters: []rune("abc")},
			expected: 6,
		},
		"word with duplicate letters": {
			configuration: game.Configuration{
				Letters: []game.LetterConfiguration{
					{
						Letter: 'a',
						Points: 1,
					},
					{
						Letter: 'b',
						Points: 2,
					},
					{
						Letter: 'c',
						Points: 3,
					},
				},
			},
			word:     game.Word{Letters: []rune("abcc")},
			expected: 9,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			score := test.configuration.GetScoreForWord(test.word)
			assert.Equal(t, test.expected, score)
		})

	}
}
