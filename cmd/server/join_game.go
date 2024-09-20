package server

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
)

type (
	joinGameRequest struct {
		PlayerName string `json:"playerName"`
		GameID     string `json:"gameId"`
	}

	joinGameResponse struct {
		GameID       string         `json:"gameId"`
		PlayerID     string         `json:"playerId"`
		Players      []playerInfo   `json:"players"`
		LetterPoints map[string]int `json:"letterPoints"`
	}

	playerInfo struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	newPlayerResponse struct {
		PlayerID string `json:"playerId"`
	}
)

func (req joinGameRequest) Execute(server *Server, conn *websocket.Conn) error {
	game, err := server.store.GetGameByID(context.Background(), req.GameID)
	if err != nil {
		return fmt.Errorf("error getting game: %w", err)
	}
	if game == nil {
		return fmt.Errorf("game not found: %s", req.GameID)
	}

	player, err := game.AddPlayer(req.PlayerName)
	if err != nil {
		return fmt.Errorf("error adding player: %w", err)
	}
	server.saveSession(conn, session{gameID: req.GameID, playerID: player.ID})

	err = server.store.SaveGame(context.Background(), game)
	if err != nil {
		return fmt.Errorf("error saving game: %w", err)
	}

	players := make([]playerInfo, len(game.Players))
	for i, p := range game.Players {
		players[i] = playerInfo{ID: p.ID, Name: p.Name}
	}

	letterPoints := make(map[string]int)
	for letter, points := range game.Config.LetterPoints {
		letterPoints[string(letter)] = points
	}

	return server.broadcastResponse(game.ID, func(s session) (string, any) {
		if s.playerID == player.ID {
			return "join_game", joinGameResponse{
				GameID:       game.ID,
				PlayerID:     player.ID,
				Players:      players,
				LetterPoints: letterPoints,
			}
		}

		return "new_player", newPlayerResponse{
			PlayerID: player.ID,
		}
	})
}
