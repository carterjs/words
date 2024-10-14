package main

import (
	"context"

	"github.com/carterjs/words/internal/words"
	"github.com/gorilla/websocket"
)

type (
	createGameRequest struct {
		PlayerName string `json:"playerName"`
	}

	createGameResponse struct {
		GameID       string                 `json:"gameId"`
		PlayerID     string                 `json:"playerId"`
		Players      []playerInfo           `json:"players"`
		LetterPoints map[string]int         `json:"letterPoints"`
		Grid         map[words.Point]string `json:"grid"`
	}
)

func (createGameRequest) Type() string {
	return "create_game"
}

func (req createGameRequest) Execute(server *Server, conn *websocket.Conn) error {
	game := words.NewGame(words.StandardConfig, req.PlayerName)
	err := server.store.SaveGame(context.Background(), game)
	if err != nil {
		return err
	}

	playerID := game.Players[0].ID

	server.saveSession(conn, session{gameID: game.ID, playerID: playerID})

	letterPoints := make(map[string]int)
	for letter, points := range game.Config.LetterPoints {
		letterPoints[string(letter)] = points
	}

	return server.sendResponse(conn, "create_game", createGameResponse{
		GameID:       game.ID,
		PlayerID:     playerID,
		Players:      []playerInfo{{ID: playerID, Name: req.PlayerName}},
		LetterPoints: letterPoints,
		Grid:         getFullGrid(game),
	})
}
