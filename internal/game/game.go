package game

import (
	"crypto"

	"github.com/google/uuid"
)

type Game struct {
	ID              string
	Name            string
	PassphraseHash  []byte
	Round           int
	ConfigurationID string
	Letters         []rune
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

func (game *Game) TakeLetters(n int) ([]rune, error) {
	if len(game.Letters) == 0 {
		return nil, ErrNoLettersInPool
	}

	if n > len(game.Letters) {
		n = len(game.Letters)
	}

	letters := game.Letters[:n]
	game.Letters = game.Letters[n:]

	return letters, nil
}
