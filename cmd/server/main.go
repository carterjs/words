package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/carterjs/words/internal/game"
	"github.com/carterjs/words/internal/store"
)

var (
	port = envOr("PORT", "8080")
)

func envOr(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	store, err := store.NewSQLite("app.db")
	if err != nil {
		panic(err)
	}

	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)

	gameService := game.NewService(logger.With("source", "gameService"), store)

	server := Server{
		Logger: logger,
		Games:  gameService,
	}

	if err := http.ListenAndServe(":"+port, server.Handler()); err != nil {
		panic(err)
	}
}
