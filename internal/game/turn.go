package game

import "github.com/google/uuid"

type Turn struct {
	ID       string
	GameID   string
	PlayerID string
	Round    int
	Word     Word
	Status   TurnStatus
}

type TurnStatus string

const (
	TurnStatusPending  TurnStatus = "PENDING"
	TurnStatusPlayed   TurnStatus = "PLAYED"
	TurnStatusRejected TurnStatus = "REJECTED"
)

func NewTurn(gameID string, round int, playerID string, word Word) *Turn {
	return &Turn{
		ID:       uuid.NewString(),
		GameID:   gameID,
		Round:    round,
		PlayerID: playerID,
		Word:     word,
		Status:   TurnStatusPending,
	}
}

type TurnVote struct {
	PlayerID string
	TurnID   string
	Value    TurnVoteValue
}

type TurnVoteValue string

const (
	TurnVoteValueApprove TurnVoteValue = "APPROVE"
	TurnVoteValueReject  TurnVoteValue = "REJECT"
)
