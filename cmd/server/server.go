package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/carterjs/words/internal/pubsub"
	"github.com/carterjs/words/internal/words"
	"net/http"
)

type (
	Server struct {
		gameStore GameStore
		events    EventBroker
	}

	GameStore interface {
		SaveGame(ctx context.Context, game *words.Game) error
		GetGameByID(ctx context.Context, id string) (*words.Game, error)
	}

	EventBroker interface {
		Publish(ctx context.Context, channel string, event GameEvent)
		Subscribe(ctx context.Context, channels ...string) (<-chan GameEvent, func())
	}

	Connection struct {
		GameID   string
		PlayerID string
	}
)

func NewServer(gameStore GameStore) *Server {
	return &Server{
		gameStore: gameStore,
		events:    pubsub.NewLocal[string, GameEvent](),
	}
}

func gameChannel(gameID string) string {
	return fmt.Sprintf("game:%s", gameID)
}

func gamePlayerChannel(gameID, playerID string) string {
	return fmt.Sprintf("game:%s:player:%s", gameID, playerID)
}

func parseRequestBody[T any](r *http.Request) (*T, error) {
	var body T
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	return &body, nil
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
