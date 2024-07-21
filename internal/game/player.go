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

func (player Player) HasLetters(letters []rune) bool {
	playerLetters := make(map[rune]int)
	for _, letter := range player.Letters {
		playerLetters[letter]++
	}

	for _, letter := range letters {
		if playerLetters[letter] == 0 {
			return false
		}
		playerLetters[letter]--
	}

	return true
}
