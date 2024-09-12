package words

import (
	"github.com/google/uuid"
)

type (
	Turn struct {
		ID       string
		GameID   string
		PlayerID string
		Round    int
		Word     Word
		Status   TurnStatus
	}

	TurnStatus string
)

const (
	TurnStatusPending TurnStatus = "PENDING"
	TurnStatusPlayed  TurnStatus = "PLAYED"
)

func newTurn(gameID string, round int, playerID string, w Word) *Turn {
	return &Turn{
		ID:       uuid.NewString(),
		GameID:   gameID,
		Round:    round,
		PlayerID: playerID,
		Word:     w,
		Status:   TurnStatusPending,
	}
}
