package words

import "encoding/json"

// EventType names a kind of game event delivered to subscribers.
type EventType string

const (
	// EventTypePlayerJoined announces a new player in the game.
	EventTypePlayerJoined EventType = "PLAYER_JOINED"
	// EventTypeGameStarted announces that the game has started.
	EventTypeGameStarted EventType = "GAME_STARTED"
	// EventTypeWordPlayed announces a word placed on the board.
	EventTypeWordPlayed EventType = "WORD_PLAYED"
	// EventTypeTurnPassed announces a forfeited turn.
	EventTypeTurnPassed EventType = "TURN_PASSED"
	// EventTypeLettersExchanged announces a player swapping letters.
	EventTypeLettersExchanged EventType = "LETTERS_EXCHANGED"
	// EventTypeRackUpdated carries a player's new rack on their private channel.
	EventTypeRackUpdated EventType = "RACK_UPDATED"
	// EventTypeChallengeStarted announces a vote on the last played word.
	EventTypeChallengeStarted EventType = "CHALLENGE_STARTED"
	// EventTypeChallengeVoteCast announces a vote added to the open challenge.
	EventTypeChallengeVoteCast EventType = "CHALLENGE_VOTE_CAST"
	// EventTypeChallengeResolved announces the outcome of a challenge.
	EventTypeChallengeResolved EventType = "CHALLENGE_RESOLVED"
	// EventTypeGameEnded announces the end of the game and final scores.
	EventTypeGameEnded EventType = "GAME_ENDED"
)

// Event is a notification about a change to a game. Payload is the JSON
// encoding of the event's payload type.
type Event struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// PlayerJoinedPayload is the payload of EventTypePlayerJoined.
type PlayerJoinedPayload struct {
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
}

// GameStartedPayload is the payload of EventTypeGameStarted. Letters is only
// populated on a player's private channel.
type GameStartedPayload struct {
	Letters []string `json:"letters"`
}

// WordPlayedPayload is the payload of EventTypeWordPlayed.
type WordPlayedPayload struct {
	PlayerID     string    `json:"playerId"`
	X            int       `json:"x"`
	Y            int       `json:"y"`
	Direction    Direction `json:"direction"`
	Word         string    `json:"word"`
	Points       int       `json:"points"`
	NextPlayerID string    `json:"nextPlayerId"`
	Round        int       `json:"round"`
}

// TurnPassedPayload is the payload of EventTypeTurnPassed.
type TurnPassedPayload struct {
	PlayerID     string `json:"playerId"`
	NextPlayerID string `json:"nextPlayerId"`
	Round        int    `json:"round"`
}

// LettersExchangedPayload is the payload of EventTypeLettersExchanged. The
// letters themselves stay private; only the count is public.
type LettersExchangedPayload struct {
	PlayerID     string `json:"playerId"`
	Count        int    `json:"count"`
	NextPlayerID string `json:"nextPlayerId"`
	Round        int    `json:"round"`
}

// RackUpdatedPayload is the payload of EventTypeRackUpdated.
type RackUpdatedPayload struct {
	Letters []string `json:"letters"`
}

// ChallengeStartedPayload is the payload of EventTypeChallengeStarted.
type ChallengeStartedPayload struct {
	ChallengerID   string `json:"challengerId"`
	MoverID        string `json:"moverId"`
	VotesInvalid   int    `json:"votesInvalid"`
	VotesValid     int    `json:"votesValid"`
	VotesNeeded    int    `json:"votesNeeded"`
	EligibleVoters int    `json:"eligibleVoters"`
}

// ChallengeVoteCastPayload is the payload of EventTypeChallengeVoteCast. The
// voter is named but their choice stays private; only the tally is public.
type ChallengeVoteCastPayload struct {
	PlayerID     string `json:"playerId"`
	VotesInvalid int    `json:"votesInvalid"`
	VotesValid   int    `json:"votesValid"`
	VotesNeeded  int    `json:"votesNeeded"`
}

// ChallengeResolvedPayload is the payload of EventTypeChallengeResolved.
type ChallengeResolvedPayload struct {
	Upheld        bool   `json:"upheld"`
	ChallengerID  string `json:"challengerId"`
	MoverID       string `json:"moverId"`
	VotesInvalid  int    `json:"votesInvalid"`
	VotesValid    int    `json:"votesValid"`
	RescindedWord string `json:"rescindedWord,omitempty"`
}

// GameEndedPayload is the payload of EventTypeGameEnded.
type GameEndedPayload struct {
	WinnerIDs []string       `json:"winnerIds"`
	Scores    map[string]int `json:"scores"`
}
