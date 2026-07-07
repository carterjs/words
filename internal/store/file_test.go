package store_test

import (
	"testing"

	"github.com/carterjs/words/internal/store"
	"github.com/carterjs/words/internal/words"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFS_GameByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		save    bool
		wantErr error
	}{
		{name: "roundtrips a saved game", save: true},
		{name: "reports a missing game", wantErr: words.ErrGameNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fileStore := store.NewFS(t.TempDir())

			game := newSavableGame(t)
			if test.save {
				require.NoError(t, fileStore.SaveGame(t.Context(), game))
			}

			loaded, err := fileStore.GameByID(t.Context(), game.ID())

			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, game.State(), loaded.State())
		})
	}
}

// newSavableGame builds a started game with a word on the board so the
// roundtrip covers players, racks, and board replay.
func newSavableGame(t *testing.T) *words.Game {
	t.Helper()

	game := words.NewGame(words.Config{
		LetterDistribution: map[rune]int{'A': 10},
		LetterPoints:       map[rune]int{'A': 1},
		RackSize:           3,
	})

	_, err := game.AddPlayer("player-0")
	require.NoError(t, err)
	require.NoError(t, game.Start())

	word := words.NewWord(words.NewPoint(0, 0), words.DirectionHorizontal, "AA")
	_, err = game.PlayWord(game.CurrentPlayerID(), word)
	require.NoError(t, err)

	return game
}
