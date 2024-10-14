package main

import (
	"context"
	"fmt"

	"github.com/carterjs/words/internal/words"
	"github.com/gorilla/websocket"
)

type (
	startGameRequest struct{}

	startGameResponse struct {
		Turn int                    `json:"turn"`
		Grid map[words.Point]string `json:"grid"`
		Rack []string               `json:"rack"`
	}
)

func (req startGameRequest) Execute(server *Server, conn *websocket.Conn) error {
	s := server.getSession(conn)
	if s.gameID == "" {
		return fmt.Errorf("no game to start")
	}

	game, err := server.store.GetGameByID(context.Background(), s.gameID)
	if err != nil {
		return fmt.Errorf("error getting game: %w", err)
	}
	if game == nil {
		return fmt.Errorf("game not found: %s", s.gameID)
	}

	if err := game.Start(); err != nil {
		return fmt.Errorf("error starting game: %w", err)
	}

	if err := server.store.SaveGame(context.Background(), game); err != nil {
		return fmt.Errorf("error saving game: %w", err)
	}

	return server.broadcastResponse(game.ID, func(sess session) (string, any) {
		var letterRack []string
		for _, letter := range game.GetPlayerByID(sess.playerID).Letters {
			letterRack = append(letterRack, string(letter))
		}

		return "start_game", startGameResponse{
			Turn: game.Turn,
			Grid: getFullGrid(game),
			Rack: letterRack,
		}
	})
}
