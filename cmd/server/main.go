// Package main runs the words HTTP server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/carterjs/words/internal/api"
	"github.com/carterjs/words/internal/pubsub"
	"github.com/carterjs/words/internal/store"
	"github.com/carterjs/words/internal/words"
)

const (
	// cleanupInterval is how often idle games are swept.
	cleanupInterval = time.Hour
	// maxGameIdle is how long an unfinished game may go without a save
	// before it is deleted. Finished games are kept.
	maxGameIdle = 14 * 24 * time.Hour
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	port := envOrDefault("PORT", "8080")

	fileStore := store.NewFS(envOrDefault("DATA_DIR", "/tmp/word-game"))
	go removeIdleGames(fileStore, logger)

	service := words.NewService(
		fileStore,
		pubsub.NewGameBroker(),
		logger,
	)

	server := api.NewServer(service, logger, api.Config{
		PublicDirectory: envOrDefault("PUBLIC_DIR", ""),
		AllowedOrigin:   envOrDefault("ALLOWED_ORIGIN", ""),
	})

	logger.Info("starting server", "port", port)
	if err := http.ListenAndServe(":"+port, server.Handler()); err != nil {
		panic(fmt.Sprintf("server stopped: %v", err))
	}
}

// removeIdleGames periodically deletes unfinished games that have gone stale.
func removeIdleGames(fileStore *store.FS, logger *slog.Logger) {
	for {
		removed, err := fileStore.RemoveIdleGames(context.Background(), maxGameIdle)
		if err != nil {
			logger.Error("cleaning up idle games", "error", err)
		} else if removed > 0 {
			logger.Info("cleaned up idle games", "count", removed)
		}

		time.Sleep(cleanupInterval)
	}
}

func envOrDefault(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return fallback
}
