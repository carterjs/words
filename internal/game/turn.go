package game

import "github.com/google/uuid"

type Turn struct {
	ID       string
	GameID   string
	PlayerID string
	Round    int
	Word     Word
	Points   string
}

func NewTurn(gameID string, playerID string, word Word) *Turn {
	return &Turn{
		ID:       uuid.NewString(),
		GameID:   gameID,
		PlayerID: playerID,
		Word:     word,
	}
}
