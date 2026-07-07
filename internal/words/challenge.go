package words

// Vote is a player's judgment of a challenged word.
type Vote string

const (
	// VoteInvalid votes to remove the challenged word from the board.
	VoteInvalid Vote = "INVALID"
	// VoteValid votes to let the challenged word stand.
	VoteValid Vote = "VALID"
)

// ChallengeOutcome describes the state of a challenge after it is opened or
// voted on. There is no dictionary: validity is decided by consensus of the
// players. A challenge resolves as soon as its outcome is mathematically
// decided; upholding it requires a strict majority of the eligible voters
// (every player except the one who played the word).
type ChallengeOutcome struct {
	ChallengerID   string
	MoverID        string
	VotesInvalid   int
	VotesValid     int
	VotesNeeded    int
	EligibleVoters int
	Resolved       bool
	Upheld         bool
	RescindedWord  *Word
}

// challengeRecord tracks an open vote on the last played word.
type challengeRecord struct {
	challengerID string
	votes        map[string]Vote
}
