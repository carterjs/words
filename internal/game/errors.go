package game

import (
	"errors"
	"fmt"
)

var (
	ErrGameNotFound          = errors.New("game not found")
	ErrPlayerNotFound        = errors.New("player not found")
	ErrConfigurationNotFound = errors.New("configuration not found")
	ErrTurnNotFound          = errors.New("turn not found")
	ErrIncorrectPassphrase   = errors.New("incorrect passphrase")
	ErrWordNotConnected      = errors.New("word is not connected to any other words")
	ErrUnchangedBoard        = errors.New("no words were added to the board")
	ErrNoLettersInPool       = errors.New("no letters in pool")
	ErrSelfVote              = errors.New("player cannot vote for themselves")
)

type WordConflictError struct {
	X, Y int
	Want rune
	Got  rune
}

func (e WordConflictError) Error() string {
	return fmt.Sprintf("conflict at (%d, %d): want %q, got %q", e.X, e.Y, e.Want, e.Got)
}
