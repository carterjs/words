package store_test

import (
	"context"
	"github.com/carterjs/words/internal/store"
	"github.com/carterjs/words/internal/words"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFS(t *testing.T) {
	t.Parallel()

	t.Run("successful store and retrieve", func(t *testing.T) {
		t.Parallel()

		fs := store.NewFS(t.TempDir())

		config := words.Presets[0].Config
		game := &words.Game{
			Started:   false,
			ID:        "gameId",
			Round:     1,
			Config:    config,
			Pool:      []rune{'A', 'B', 'C'},
			PoolIndex: 0,
			Players: []words.Player{
				{
					ID:     "playerId",
					GameID: "gameId",
					Name:   "playerName",
				},
			},
			Turn:  0,
			Board: words.NewBoard("gameId", config),
		}

		err := fs.SaveGame(context.Background(), game)
		assert.NoError(t, err)

		g, err := fs.GetGameByID(context.Background(), game.ID)
		assert.NoError(t, err)
		assert.Equal(t, game, g)
	})

	t.Run("game not found", func(t *testing.T) {
		t.Parallel()

		fs := store.NewFS(t.TempDir())

		g, err := fs.GetGameByID(context.Background(), "unknown")
		assert.ErrorIs(t, err, nil)
		assert.Nil(t, g)
	})
}
