package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/carterjs/words/internal/accesslog"
	"github.com/carterjs/words/internal/game"
)

type (
	Server struct {
		Logger *slog.Logger
		Games  Games
	}

	Games interface {
		CreateGame(ctx context.Context, name, passphrase string) (*game.Game, error)
		GetGameByID(ctx context.Context, id string) (*game.Game, error)
		AddPlayerToGame(ctx context.Context, gameID, name, passphrase string) (*game.Player, error)
		GetPlayersByGameID(ctx context.Context, gameID string) ([]game.Player, error)
	}
)

func (server *Server) Handler() http.HandlerFunc {
	mux := http.NewServeMux()

	mux.Handle("/health", server.handleGetHealth())

	mux.Handle("POST /v1/games", server.handleCreateGame())
	mux.Handle("GET /v1/games/{gameId}", server.handleGetGame())
	mux.Handle("POST /v1/games/{gameId}/players", server.handleAddPlayerToGame())
	mux.Handle("GET /v1/games/{gameId}/players", server.handleGetGamePlayers())
	// mux.Handle("POST /v1/games/{gameId}/rounds/{roundId}/turns", server.handleRecordTurn())

	return accesslog.NewMiddleware(server.Logger, mux)
}

func respondWithJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func (server *Server) handleGetHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
