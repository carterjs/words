package main

import (
	"github.com/carterjs/words/internal/store"
	"log/slog"
	"net/http"
	"os"
)

var (
	port      = envOrDefault("PORT", "8080")
	dataDir   = envOrDefault("DATA_DIR", "/tmp/word-game")
	publicDir = envOrDefault("PUBLIC_DIR", "public")
)

func envOrDefault(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	server := NewServer(store.NewFS(dataDir))

	slog.Info("Starting server", "port", port)
	if err := http.ListenAndServe(":"+port, server.Handler()); err != nil {
		panic(err)
	}
}
