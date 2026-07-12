// Package errcode names the words service's errors as stable,
// transport-neutral codes. Each code carries a class — a coarse category the
// presentation layer maps to a transport status — and a description. It
// imports the service packages, never the reverse, and knows nothing about
// HTTP.
package errcode

import (
	"errors"

	"github.com/carterjs/words/internal/words"
)

// Class is a coarse, transport-neutral category shared by many codes.
type Class string

const (
	// ClassInvalid marks requests that break validation or game rules.
	ClassInvalid Class = "invalid"
	// ClassNotFound marks requests for things that do not exist.
	ClassNotFound Class = "not_found"
	// ClassConflict marks requests that clash with the game's current state.
	ClassConflict Class = "conflict"
	// ClassUnauthenticated marks requests with no player identity.
	ClassUnauthenticated Class = "unauthenticated"
	// ClassInternal marks unexpected failures on our side.
	ClassInternal Class = "internal"
)

// Code is a stable string identifier for one error condition.
type Code string

type definition struct {
	class       Class
	description string
}

var definitions = make(map[Code]definition)

func define(name string, class Class, description string) Code {
	code := Code(name)
	definitions[code] = definition{class: class, description: description}

	return code
}

// Codes returned by the words service.
var (
	// Unknown covers errors with no more specific code.
	Unknown = define("unknown", ClassInternal, "an unexpected error occurred")
	// GameNotFound reports a request for a game that does not exist.
	GameNotFound = define("game_not_found", ClassNotFound, "the requested game does not exist")
	// PresetNotFound reports a request for a preset that does not exist.
	PresetNotFound = define("preset_not_found", ClassNotFound, "the requested preset does not exist")
	// PlayerNotFound reports a player who is not part of the game.
	PlayerNotFound = define("player_not_found", ClassNotFound, "the player is not part of this game")
	// GameNotStarted reports an action that requires a started game.
	GameNotStarted = define("game_not_started", ClassConflict, "the game has not started yet")
	// GameAlreadyStarted reports an action that requires an unstarted game.
	GameAlreadyStarted = define("game_already_started", ClassConflict, "the game has already started")
	// GameFinished reports an action against a finished game.
	GameFinished = define("game_finished", ClassConflict, "the game is over")
	// NotYourTurn reports an action taken out of turn.
	NotYourTurn = define("not_your_turn", ClassConflict, "it is another player's turn")
	// ChallengePending reports an action blocked by an open challenge vote.
	ChallengePending = define("challenge_pending", ClassConflict, "a challenge vote must resolve first")
	// NoPendingChallenge reports a vote with no challenge open.
	NoPendingChallenge = define("no_pending_challenge", ClassConflict, "there is no open challenge to vote on")
	// NothingToChallenge reports a challenge with no unsettled word.
	NothingToChallenge = define("nothing_to_challenge", ClassConflict, "there is no word to challenge")
	// CannotChallengeOwnWord reports a player challenging their own word.
	CannotChallengeOwnWord = define("cannot_challenge_own_word", ClassInvalid, "you cannot challenge your own word")
	// CannotVoteOnOwnWord reports the word's player trying to vote.
	CannotVoteOnOwnWord = define("cannot_vote_on_own_word", ClassInvalid, "you cannot vote on your own word")
	// AlreadyVoted reports a repeated vote on the same challenge.
	AlreadyVoted = define("already_voted", ClassConflict, "you already voted on this challenge")
	// InvalidVote reports an unrecognized vote value.
	InvalidVote = define("invalid_vote", ClassInvalid, "votes must be VALID or INVALID")
	// NotEnoughPlayers reports starting a game without players.
	NotEnoughPlayers = define("not_enough_players", ClassInvalid, "the game needs at least one player")
	// CannotPlayWord reports a word the player lacks the letters for.
	CannotPlayWord = define("cannot_play_word", ClassInvalid, "you do not have the letters to play that word")
	// WordNotConnected reports a word that touches no existing word.
	WordNotConnected = define("word_not_connected", ClassInvalid, "the word must connect to an existing word")
	// FirstWordNotCentered reports an opening word missing the center.
	FirstWordNotCentered = define("first_word_not_centered", ClassInvalid, "the first word must cross the center")
	// WordIncomplete reports a word running into adjacent letters.
	WordIncomplete = define("word_incomplete", ClassInvalid, "the word runs into adjacent letters")
	// WordUnchanged reports a placement that adds no letters.
	WordUnchanged = define("word_unchanged", ClassInvalid, "the placement adds no new letters")
	// WordConflict reports a placement that disagrees with the board.
	WordConflict = define("word_conflict", ClassInvalid, "the word conflicts with letters already on the board")
	// MissingLetters reports an exchange of letters the player lacks.
	MissingLetters = define("missing_letters", ClassInvalid, "you do not have those letters")
	// NotEnoughLettersInPool reports an exchange larger than the pool.
	NotEnoughLettersInPool = define("not_enough_letters_in_pool", ClassInvalid, "the pool does not have enough letters")
	// BadRequest reports a request body or parameter that could not be parsed.
	BadRequest = define("bad_request", ClassInvalid, "the request could not be parsed")
	// UnknownOperation reports an update operation the API does not know.
	UnknownOperation = define("unknown_operation", ClassInvalid, "unknown operation")
	// MissingPlayer reports a request that requires a player identity.
	MissingPlayer = define("missing_player", ClassUnauthenticated, "the request has no player identity")
)

// Class returns the code's category.
func (code Code) Class() Class {
	if definition, exists := definitions[code]; exists {
		return definition.class
	}

	return ClassInternal
}

// Description returns the human-readable meaning of the code.
func (code Code) Description() string {
	if definition, exists := definitions[code]; exists {
		return definition.description
	}

	return definitions[Unknown].description
}

// FromError maps a service error to its code.
func FromError(err error) Code {
	var conflict words.WordConflictError
	if errors.As(err, &conflict) {
		return WordConflict
	}

	for sentinel, code := range sentinelCodes {
		if errors.Is(err, sentinel) {
			return code
		}
	}

	return Unknown
}

var sentinelCodes = map[error]Code{
	words.ErrGameNotFound:           GameNotFound,
	words.ErrPresetNotFound:         PresetNotFound,
	words.ErrPlayerNotFound:         PlayerNotFound,
	words.ErrGameNotStarted:         GameNotStarted,
	words.ErrGameStarted:            GameAlreadyStarted,
	words.ErrGameFinished:           GameFinished,
	words.ErrNotYourTurn:            NotYourTurn,
	words.ErrChallengePending:       ChallengePending,
	words.ErrNoPendingChallenge:     NoPendingChallenge,
	words.ErrNothingToChallenge:     NothingToChallenge,
	words.ErrCannotChallengeOwnWord: CannotChallengeOwnWord,
	words.ErrCannotVoteOnOwnWord:    CannotVoteOnOwnWord,
	words.ErrAlreadyVoted:           AlreadyVoted,
	words.ErrInvalidVote:            InvalidVote,
	words.ErrNotEnoughPlayers:       NotEnoughPlayers,
	words.ErrCannotPlayWord:         CannotPlayWord,
	words.ErrWordNotConnected:       WordNotConnected,
	words.ErrFirstWordNotCentered:   FirstWordNotCentered,
	words.ErrIncomplete:             WordIncomplete,
	words.ErrUnchanged:              WordUnchanged,
	words.ErrMissingLetters:         MissingLetters,
	words.ErrNotEnoughLettersInPool: NotEnoughLettersInPool,
}
