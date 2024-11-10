package words

import (
	"errors"
	"fmt"
)

var (
	ErrPlayerNotFound       = errors.New("player not found")
	ErrCannotPlayWord       = errors.New("player cannot play word")
	ErrWordNotConnected     = errors.New("word is not connected to any other words")
	ErrUnchanged            = errors.New("no words were added to the board")
	ErrNoLettersInPool      = errors.New("no letters in pool")
	ErrNotEnoughPlayers     = errors.New("not enough players")
	ErrFirstWordNotCentered = errors.New("first word must be centered")
	ErrNothingToUndo        = errors.New("nothing to undo")
	ErrGameNotStarted       = errors.New("game not started")
	ErrNotYourTurn          = errors.New("not your turn")
	ErrGameStarted          = errors.New("game already started")
	ErrIncomplete           = errors.New("word is incomplete")
)

type WordConflictError struct {
	X, Y int
	Want rune
	Got  rune
}

func (e WordConflictError) Error() string {
	return fmt.Sprintf("conflict at (%d, %d): want %q, got %q", e.X, e.Y, e.Want, e.Got)
}
