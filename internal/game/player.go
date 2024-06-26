package game

import "github.com/google/uuid"

type Player struct {
	ID      string
	GameID  string
	Name    string
	Status  PlayerStatus
	Letters []rune
}

type PlayerStatus string

const (
	PlayerStatusActive   PlayerStatus = "active"
	PlayerStatusInactive PlayerStatus = "inactive"
)

func NewPlayer(gameID string, name string) *Player {
	return &Player{
		ID:     uuid.NewString(),
		GameID: gameID,
		Name:   name,
		Status: PlayerStatusActive,
	}
}

func (player Player) HasLettersForWord(word string) bool {
	letters := make(map[rune]int)
	for _, l := range player.Letters {
		letters[l]++
	}

	for _, l := range word {
		if letters[l] == 0 {
			return false
		}
		letters[l]--
	}

	return true
}
