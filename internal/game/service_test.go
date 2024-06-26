package game_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/carterjs/words/internal/game"
	"github.com/carterjs/words/internal/store"
	"github.com/stretchr/testify/assert"
)

func TestGameService(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	store, err := store.NewSQLite(filepath.Join(t.TempDir(), "test.db"))
	assert.NoError(t, err)

	ctx := context.Background()
	gameService := game.NewService(logger, store)
	gameService.CreateGame(ctx, "game1", "passphrase")

}
