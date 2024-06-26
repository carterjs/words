package game

import (
	"crypto"

	"github.com/google/uuid"
)

type Game struct {
	ID             string
	Name           string
	PassphraseHash []byte
	Round          int
}

func New(name string, passphrase string) *Game {
	return &Game{
		ID:             uuid.NewString(),
		Name:           name,
		PassphraseHash: HashPassphrase(passphrase),
		Round:          1,
	}
}

func (game Game) PassphraseMatches(input string) bool {
	// check against hash passphrase hash
	inputHash := HashPassphrase(input)
	return string(inputHash) == string(game.PassphraseHash)
}

func HashPassphrase(passphrase string) []byte {
	hash := crypto.SHA256.New()
	hash.Write([]byte(passphrase))
	return hash.Sum(nil)
}
