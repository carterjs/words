package main

import (
	"context"
	"fmt"

	"github.com/carterjs/words/internal/words"
	"github.com/gorilla/websocket"
)

type (
	playWordRequest struct {
		X         int    `json:"x"`
		Y         int    `json:"y"`
		Direction string `json:"direction"`
		Word      string `json:"word"`
	}

	playWordResponse struct {
		Word    string                 `json:"word"`
		Grid    map[words.Point]string `json:"grid"`
		Points  int                    `json:"points"`
		NewRack []string               `json:"rack"`
	}

	newWordResponse struct {
		Word   string                 `json:"word"`
		Grid   map[words.Point]string `json:"grid"`
		Points int                    `json:"points"`
	}
)

const playWordResponseType = "play_word"

func (req playWordRequest) Execute(server *Server, conn *websocket.Conn) error {
	s := server.getSession(conn)
	if s.gameID == "" {
		return fmt.Errorf("no game to play word in")
	}

	game, err := server.store.GetGameByID(context.Background(), s.gameID)
	if err != nil {
		return fmt.Errorf("error getting game: %w", err)
	}
	if game == nil {
		return fmt.Errorf("game not found: %s", s.gameID)
	}

	direction := words.DirectionHorizontal
	if req.Direction == "vertical" {
		direction = words.DirectionVertical
	}

	word := words.NewWord(words.NewPoint(req.X, req.Y), direction, req.Word)
	result, err := game.PlayWord(s.playerID, word)
	if err != nil {
		return fmt.Errorf("error playing word: %w", err)
	}

	if err := server.store.SaveGame(context.Background(), game); err != nil {
		return fmt.Errorf("error saving game: %w", err)
	}

	playerID := s.playerID
	last, _, _ := word.Index(len(word.Letters) - 1)
	newRack := make([]string, len(game.GetPlayerByID(s.playerID).Letters))
	for i, letter := range game.GetPlayerByID(s.playerID).Letters {
		newRack[i] = string(letter)
	}

	return server.broadcastResponse(s.gameID, func(s session) (string, any) {
		if s.playerID == playerID {
			return playWordResponseType, playWordResponse{
				Word:    string(word.Letters),
				Grid:    getPartialGrid(game, word.Start.X(), word.Start.Y(), last.X(), last.Y()),
				Points:  result.Points,
				NewRack: newRack,
			}
		}

		return playWordResponseType, newWordResponse{
			Word:   string(word.Letters),
			Grid:   getPartialGrid(game, word.Start.X(), word.Start.Y(), last.X(), last.Y()),
			Points: result.Points,
		}
	})
}
