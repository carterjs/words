// Package main runs the words HTTP server.
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/carterjs/words/internal/api"
	"github.com/carterjs/words/internal/pubsub"
	"github.com/carterjs/words/internal/store"
	"github.com/carterjs/words/internal/words"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	port := envOrDefault("PORT", "8080")

	service := words.NewService(
		store.NewFS(envOrDefault("DATA_DIR", "/tmp/word-game")),
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

func envOrDefault(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return fallback
}
