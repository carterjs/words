package words

import (
	"errors"
	"fmt"
)

var (
	// ErrGameNotFound reports that no game exists with the requested ID.
	ErrGameNotFound = errors.New("game not found")
	// ErrPresetNotFound reports that no preset exists with the requested ID.
	ErrPresetNotFound = errors.New("preset not found")
	// ErrPlayerNotFound reports that the player is not part of the game.
	ErrPlayerNotFound = errors.New("player not found")
	// ErrCannotPlayWord reports that the player lacks the letters to play the word.
	ErrCannotPlayWord = errors.New("player cannot play word")
	// ErrWordNotConnected reports that the word does not touch any existing word.
	ErrWordNotConnected = errors.New("word is not connected to any other words")
	// ErrUnchanged reports that the placement would not add any new letters.
	ErrUnchanged = errors.New("no letters were added to the board")
	// ErrNoLettersInPool reports that the letter pool is exhausted.
	ErrNoLettersInPool = errors.New("no letters in pool")
	// ErrNotEnoughLettersInPool reports that the pool cannot cover an exchange.
	ErrNotEnoughLettersInPool = errors.New("not enough letters in pool")
	// ErrMissingLetters reports that the player does not hold the letters offered.
	ErrMissingLetters = errors.New("player does not have those letters")
	// ErrNotEnoughPlayers reports that the game cannot start without players.
	ErrNotEnoughPlayers = errors.New("not enough players")
	// ErrFirstWordNotCentered reports that the opening word misses the center cell.
	ErrFirstWordNotCentered = errors.New("first word must be centered")
	// ErrGameNotStarted reports that the action requires a started game.
	ErrGameNotStarted = errors.New("game not started")
	// ErrNotYourTurn reports that another player has the current turn.
	ErrNotYourTurn = errors.New("not your turn")
	// ErrGameStarted reports that the action requires a game that has not started.
	ErrGameStarted = errors.New("game already started")
	// ErrGameFinished reports that the game is over and cannot be changed.
	ErrGameFinished = errors.New("game is finished")
	// ErrIncomplete reports that the word runs into adjacent letters, forming a
	// longer word than the one submitted.
	ErrIncomplete = errors.New("word is incomplete")
	// ErrChallengePending reports that a challenge vote must resolve before play continues.
	ErrChallengePending = errors.New("a challenge is pending")
	// ErrNoPendingChallenge reports that there is no challenge to vote on.
	ErrNoPendingChallenge = errors.New("no pending challenge")
	// ErrNothingToChallenge reports that there is no unsettled word to challenge.
	ErrNothingToChallenge = errors.New("nothing to challenge")
	// ErrCannotChallengeOwnWord reports that a player challenged their own word.
	ErrCannotChallengeOwnWord = errors.New("cannot challenge your own word")
	// ErrAlreadyVoted reports that the player already voted on the challenge.
	ErrAlreadyVoted = errors.New("player already voted")
	// ErrCannotVoteOnOwnWord reports that the word's player tried to vote.
	ErrCannotVoteOnOwnWord = errors.New("cannot vote on your own word")
	// ErrInvalidVote reports a vote value that is neither valid nor invalid.
	ErrInvalidVote = errors.New("invalid vote")
)

// WordConflictError reports a placement that disagrees with a letter already
// on the board.
type WordConflictError struct {
	column, row int
	want, got   rune
}

// Error implements the error interface.
func (conflict WordConflictError) Error() string {
	return fmt.Sprintf("conflict at (%d, %d): want %q, got %q", conflict.column, conflict.row, conflict.want, conflict.got)
}
