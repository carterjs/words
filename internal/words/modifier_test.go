package words_test

import (
	"testing"

	"github.com/carterjs/words/internal/words"
	"github.com/stretchr/testify/assert"
)

func TestModifier_ModifyLetterScore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		modifier words.Modifier
		score    int
		want     int
	}{
		{name: "doubles a letter", modifier: words.ModifierDoubleLetter, score: 3, want: 6},
		{name: "triples a letter", modifier: words.ModifierTripleLetter, score: 3, want: 9},
		{name: "leaves letters alone for word modifiers", modifier: words.ModifierDoubleWord, score: 3, want: 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, test.modifier.ModifyLetterScore(test.score))
		})
	}
}

func TestModifier_ModifyWordScore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		modifier words.Modifier
		score    int
		want     int
	}{
		{name: "doubles a word", modifier: words.ModifierDoubleWord, score: 5, want: 10},
		{name: "triples a word", modifier: words.ModifierTripleWord, score: 5, want: 15},
		{name: "leaves words alone for letter modifiers", modifier: words.ModifierTripleLetter, score: 5, want: 5},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, test.modifier.ModifyWordScore(test.score))
		})
	}
}
