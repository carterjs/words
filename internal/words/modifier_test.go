package words_test

import (
	"github.com/carterjs/words/internal/words"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		modifier            words.Modifier
		wordScore           int
		expectedWordScore   int
		letterScore         int
		expectedLetterScore int
	}{
		{
			wordScore:           10,
			expectedWordScore:   10,
			letterScore:         1,
			expectedLetterScore: 1,
		},
		{
			modifier:            words.ModifierDoubleWord,
			wordScore:           10,
			expectedWordScore:   20,
			letterScore:         1,
			expectedLetterScore: 1,
		},
		{
			modifier:            words.ModifierTripleWord,
			wordScore:           10,
			expectedWordScore:   30,
			letterScore:         1,
			expectedLetterScore: 1,
		},
		{
			modifier:            words.ModifierDoubleLetter,
			wordScore:           10,
			expectedWordScore:   10,
			letterScore:         1,
			expectedLetterScore: 2,
		},
		{
			modifier:            words.ModifierTripleLetter,
			wordScore:           10,
			expectedWordScore:   10,
			letterScore:         1,
			expectedLetterScore: 3,
		},
	}

	for _, test := range tests {
		t.Run(string(test.modifier), func(t *testing.T) {
			t.Parallel()

			wordScore := test.modifier.ModifyWordScore(test.wordScore)
			assert.Equal(t, test.expectedWordScore, wordScore)

			letterScore := test.modifier.ModifyLetterScore(test.letterScore)
			assert.Equal(t, test.expectedLetterScore, letterScore)
		})
	}
}
