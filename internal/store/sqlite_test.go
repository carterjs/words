package store_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/carterjs/words/internal/game"
	"github.com/carterjs/words/internal/store"
	"github.com/stretchr/testify/assert"
)

func TestSQLite(t *testing.T) {
	f := filepath.Join(t.TempDir(), "test.db")

	db, err := store.NewSQLite(f)
	assert.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	err = db.SavePlayer(ctx, game.Player{
		ID:     "player1",
		GameID: "game1",
		Name:   "Alice",
		Status: game.PlayerStatusActive,
		Letters: []rune{
			'a', 'b', 'c', 'd', 'e',
		},
	})
	assert.NoError(t, err)

	players, err := db.GetPlayersByGameID(ctx, "game1")
	assert.NoError(t, err)
	assert.Len(t, players, 1)
	assert.Equal(t, "player1", players[0].ID)
	assert.Equal(t, "Alice", players[0].Name)
	assert.Equal(t, game.PlayerStatusActive, players[0].Status)
	assert.Equal(t, []rune{'a', 'b', 'c', 'd', 'e'}, players[0].Letters)
}
