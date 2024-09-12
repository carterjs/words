package words

import (
	"errors"
	"fmt"
)

var (
	ErrGameNotFound         = errors.New("game not found")
	ErrPlayerNotFound       = errors.New("player not found")
	ErrTurnNotFound         = errors.New("turn not found")
	ErrIncorrectPassphrase  = errors.New("incorrect passphrase")
	ErrCannotPlayWord       = errors.New("player cannot play word")
	ErrWordNotConnected     = errors.New("word is not connected to any other words")
	ErrUnchanged            = errors.New("no words were added to the board")
	ErrNoLettersInPool      = errors.New("no letters in pool")
	ErrPresetNotFound       = errors.New("preset not found")
	ErrBoardNotFound        = errors.New("board not found")
	ErrNotEnoughPlayers     = errors.New("not enough players")
	ErrGameAlreadyStarted   = errors.New("game already started")
	ErrFirstWordNotCentered = errors.New("first word must be centered")
	ErrNotPlayersTurn       = errors.New("not player's turn")
	ErrNothingToUndo        = errors.New("nothing to undo")
)

type WordConflictError struct {
	X, Y int
	Want rune
	Got  rune
}

func (e WordConflictError) Error() string {
	return fmt.Sprintf("conflict at (%d, %d): want %q, got %q", e.X, e.Y, e.Want, e.Got)
}
