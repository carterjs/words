package server

import (
	"context"
	"github.com/carterjs/words/internal/words"
	"github.com/gorilla/websocket"
)

type (
	rejoinGameRequest struct {
		PlayerID string `json:"playerId"`
		GameID   string `json:"gameId"`
	}

	rejoinGameResponse struct {
		GameID       string                 `json:"gameId"`
		PlayerID     string                 `json:"playerId"`
		Turn         int                    `json:"turn"`
		Started      bool                   `json:"started"`
		Grid         map[words.Point]string `json:"grid"`
		Rack         []string               `json:"rack"`
		Players      []playerInfo           `json:"players"`
		LetterPoints map[string]int         `json:"letterPoints"`
	}
)

func (req rejoinGameRequest) Execute(server *Server, conn *websocket.Conn) error {
	game, err := server.store.GetGameByID(context.Background(), req.GameID)
	if err != nil {
		return err
	}
	if game == nil {
		return nil
	}

	player := game.GetPlayerByID(req.PlayerID)
	if player == nil {
		return nil
	}

	server.saveSession(conn, session{gameID: game.ID, playerID: player.ID})

	// TODO: dedup
	var players []playerInfo
	for _, p := range game.Players {
		players = append(players, playerInfo{ID: p.ID, Name: p.Name})
	}

	// TODO: dedup
	var rack []string
	for _, letter := range player.Letters {
		rack = append(rack, string(letter))
	}

	// TODO: dedup
	letterPoints := make(map[string]int)
	for letter, points := range game.Config.LetterPoints {
		letterPoints[string(letter)] = points
	}

	return server.broadcastResponse(game.ID, func(s session) (string, any) {
		if s.playerID == player.ID {
			return "rejoin_game", rejoinGameResponse{
				GameID:       game.ID,
				PlayerID:     player.ID,
				Turn:         game.Turn,
				Grid:         getGrid(game, -8, -8, 8, 8),
				Rack:         rack,
				Players:      players,
				Started:      game.Started,
				LetterPoints: letterPoints,
			}
		}

		return "player_online", playerStatusChange{
			PlayerID: player.ID,
			Online:   true,
		}
	})
}
