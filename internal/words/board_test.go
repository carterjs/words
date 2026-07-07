package words_test

import (
	"testing"

	"github.com/carterjs/words/internal/words"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoard_PlaceWord(t *testing.T) {
	t.Parallel()

	crossed := []words.Word{horizontal(0, 0, "AB")}
	cornered := []words.Word{horizontal(0, 0, "AB"), vertical(1, 0, "BC")}

	tests := []struct {
		name              string
		setup             []words.Word
		word              words.Word
		wantErr           error
		wantConflict      bool
		wantPoints        int
		wantIndirectWords int
	}{
		{name: "rejects first word off center", word: horizontal(3, 3, "AB"), wantErr: words.ErrFirstWordNotCentered},
		{name: "places first word through center", word: horizontal(-1, 0, "ABC"), wantPoints: 6},
		{name: "rejects disconnected word", setup: crossed, word: horizontal(5, 5, "AB"), wantErr: words.ErrWordNotConnected},
		{name: "rejects conflicting overlap", setup: crossed, word: vertical(0, 0, "BC"), wantConflict: true},
		{name: "rejects placement adding no letters", setup: crossed, word: horizontal(0, 0, "AB"), wantErr: words.ErrUnchanged},
		{name: "rejects word running into an adjacent letter", setup: cornered, word: horizontal(-1, 1, "CC"), wantErr: words.ErrIncomplete},
		{name: "scores indirect words above the new word", setup: crossed, word: horizontal(0, 1, "BB"), wantPoints: 11, wantIndirectWords: 2},
		// regression: existing letters after the new letter used to loop forever
		{name: "scores indirect word extending right of the new letter", setup: crossed, word: vertical(-1, 0, "CC"), wantPoints: 12, wantIndirectWords: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// scores A=1, B=2, C=3 with no modifiers
			board := words.NewBoard(words.Config{LetterPoints: map[rune]int{'A': 1, 'B': 2, 'C': 3}})
			for _, word := range test.setup {
				_, err := board.PlaceWord(word)
				require.NoError(t, err)
			}

			result, err := board.PlaceWord(test.word)

			if test.wantConflict {
				var conflict words.WordConflictError
				assert.ErrorAs(t, err, &conflict)
				return
			}

			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.wantPoints, result.Points)
			assert.Len(t, result.IndirectWords, test.wantIndirectWords)
		})
	}
}
