package game

import (
	"errors"
	"fmt"
)

var ErrIncorrectPassphrase = errors.New("incorrect passphrase")

var ErrWordNotConnected = errors.New("word is not connected to any other words")

type WordConflictError struct {
	X, Y int
	Want rune
	Got  rune
}

func (e WordConflictError) Error() string {
	return fmt.Sprintf("conflict at (%d, %d): want %q, got %q", e.X, e.Y, e.Want, e.Got)
}
