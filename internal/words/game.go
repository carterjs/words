package words

import (
	"fmt"
	"math"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

// scorelessRoundLimit ends the game after this many consecutive full rounds
// in which every player passed or exchanged instead of playing a word.
const scorelessRoundLimit = 2

// Game is a single match: its configuration, players, letter pool, board,
// and the turn and challenge state around them.
type Game struct {
	id             string
	started        bool
	finished       bool
	round          int
	config         Config
	pool           []rune
	poolIndex      int
	players        []Player
	turn           int
	scorelessTurns int
	board          *Board
	lastWord       *lastWordRecord
	challenge      *challengeRecord
	winnerIDs      []string
}

// lastWordRecord tracks the most recently played word. A word is settled —
// no longer challengeable — once the next turn is taken or a challenge
// against it fails.
type lastWordRecord struct {
	playerID string
	settled  bool
}

// NewGame returns a new unstarted game with the given configuration.
func NewGame(config Config) *Game {
	return &Game{
		id:     uuid.NewString(),
		round:  1,
		config: config,
		pool:   initialLetterPool(config),
		board:  NewBoard(config),
	}
}

// ID returns the game's unique identifier.
func (game *Game) ID() string {
	return game.id
}

// Started reports whether the game has started.
func (game *Game) Started() bool {
	return game.started
}

// Finished reports whether the game is over.
func (game *Game) Finished() bool {
	return game.finished
}

// Round returns the current round number, starting at 1.
func (game *Game) Round() int {
	return game.round
}

// Config returns the rules the game is played with.
func (game *Game) Config() Config {
	return game.config
}

// Board returns the game's board.
func (game *Game) Board() *Board {
	return game.board
}

// Players returns the players in turn order.
func (game *Game) Players() []Player {
	players := make([]Player, len(game.players))
	copy(players, game.players)
	return players
}

// PlayerByID returns the player with the given ID and whether they are in
// the game.
func (game *Game) PlayerByID(playerID string) (Player, bool) {
	index := game.playerIndex(playerID)
	if index < 0 {
		return Player{}, false
	}

	return game.players[index], true
}

// CurrentPlayerID returns the ID of the player whose turn it is, or the
// empty string if the game has no players.
func (game *Game) CurrentPlayerID() string {
	if len(game.players) == 0 {
		return ""
	}

	return game.players[game.turn].id
}

// WinnerIDs returns the IDs of the highest-scoring players once the game is
// finished.
func (game *Game) WinnerIDs() []string {
	winnerIDs := make([]string, len(game.winnerIDs))
	copy(winnerIDs, game.winnerIDs)
	return winnerIDs
}

// LettersRemaining returns the number of letters left in the pool.
func (game *Game) LettersRemaining() int {
	return len(game.pool) - game.poolIndex
}

// AddPlayer adds a player to an unstarted game and returns them.
func (game *Game) AddPlayer(name string) (Player, error) {
	if game.started {
		return Player{}, ErrGameStarted
	}

	player := newPlayer(name)
	game.players = append(game.players, player)
	return player, nil
}

// Start begins the game, dealing every player a full rack.
func (game *Game) Start() error {
	if game.started {
		return ErrGameStarted
	}

	if len(game.players) < 1 {
		return ErrNotEnoughPlayers
	}

	for index := range game.players {
		game.fillPlayerRack(&game.players[index])
	}

	game.started = true

	return nil
}

// PlayWord places the word on the board for the given player, spends and
// replenishes their letters, and advances the turn. Word validity is not
// checked against a dictionary; opponents may challenge the word instead.
func (game *Game) PlayWord(playerID string, word Word) (PlacementResult, error) {
	if err := game.assertTurn(playerID); err != nil {
		return PlacementResult{}, fmt.Errorf("checking turn: %w", err)
	}

	result, err := game.checkWord(playerID, word)
	if err != nil {
		return PlacementResult{}, fmt.Errorf("checking word: %w", err)
	}

	result, err = game.board.PlaceWord(result.DirectWord)
	if err != nil {
		return PlacementResult{}, fmt.Errorf("placing word: %w", err)
	}

	player := &game.players[game.playerIndex(playerID)]
	player.takeLetters(lettersFromMap(result.LettersUsed))

	drawn := game.fillPlayerRack(player)
	player.turns = append(player.turns, TurnRecord{
		Points:       result.Points,
		LettersUsed:  result.LettersUsed,
		LettersDrawn: drawn,
	})

	game.settleLastWord()
	game.lastWord = &lastWordRecord{playerID: playerID}
	game.scorelessTurns = 0

	if game.LettersRemaining() == 0 && len(player.letters) == 0 {
		game.finish(playerID)
		return result, nil
	}

	game.advanceTurn()

	return result, nil
}

// PassTurn forfeits the given player's turn. The game ends when every
// player passes or exchanges for scorelessRoundLimit consecutive rounds.
func (game *Game) PassTurn(playerID string) error {
	if err := game.assertTurn(playerID); err != nil {
		return fmt.Errorf("checking turn: %w", err)
	}

	game.settleLastWord()
	game.endScorelessTurn()

	return nil
}

// ExchangeLetters swaps the given letters from the player's rack for fresh
// ones from the pool, consuming their turn.
func (game *Game) ExchangeLetters(playerID string, letters []rune) error {
	if err := game.assertTurn(playerID); err != nil {
		return fmt.Errorf("checking turn: %w", err)
	}

	if len(letters) == 0 {
		return ErrMissingLetters
	}

	if len(letters) > game.LettersRemaining() {
		return ErrNotEnoughLettersInPool
	}

	player := &game.players[game.playerIndex(playerID)]
	if !player.hasLetters(letters) {
		return ErrMissingLetters
	}

	player.takeLetters(letters)
	player.giveLetters(game.pool[game.poolIndex : game.poolIndex+len(letters)])
	game.poolIndex += len(letters)

	// return the exchanged letters to the pool and shuffle the undrawn tail
	// so they cannot be drawn back in the same order
	game.pool = append(game.pool, letters...)
	game.shufflePoolTail()

	game.settleLastWord()
	game.endScorelessTurn()

	return nil
}

// Challenge opens a vote on the last played word. The challenger implicitly
// votes that the word is invalid, which resolves the challenge immediately
// in a two-player game.
func (game *Game) Challenge(playerID string) (ChallengeOutcome, error) {
	if err := game.assertChallengeAllowed(playerID); err != nil {
		return ChallengeOutcome{}, fmt.Errorf("checking challenge: %w", err)
	}

	game.challenge = &challengeRecord{
		challengerID: playerID,
		votes:        map[string]Vote{playerID: VoteInvalid},
	}

	outcome, err := game.resolveChallenge()
	if err != nil {
		return ChallengeOutcome{}, fmt.Errorf("resolving challenge: %w", err)
	}

	return outcome, nil
}

func (game *Game) assertChallengeAllowed(playerID string) error {
	if !game.started {
		return ErrGameNotStarted
	}

	if game.finished {
		return ErrGameFinished
	}

	if game.challenge != nil {
		return ErrChallengePending
	}

	if game.lastWord == nil || game.lastWord.settled {
		return ErrNothingToChallenge
	}

	if game.playerIndex(playerID) < 0 {
		return ErrPlayerNotFound
	}

	if game.lastWord.playerID == playerID {
		return ErrCannotChallengeOwnWord
	}

	return nil
}

// CastVote records the given player's vote on the pending challenge and
// resolves the challenge once the outcome is decided.
func (game *Game) CastVote(playerID string, vote Vote) (ChallengeOutcome, error) {
	if err := game.assertVoteAllowed(playerID, vote); err != nil {
		return ChallengeOutcome{}, fmt.Errorf("checking vote: %w", err)
	}

	game.challenge.votes[playerID] = vote

	outcome, err := game.resolveChallenge()
	if err != nil {
		return ChallengeOutcome{}, fmt.Errorf("resolving challenge: %w", err)
	}

	return outcome, nil
}

func (game *Game) assertVoteAllowed(playerID string, vote Vote) error {
	if !game.started {
		return ErrGameNotStarted
	}

	if game.challenge == nil {
		return ErrNoPendingChallenge
	}

	if vote != VoteValid && vote != VoteInvalid {
		return ErrInvalidVote
	}

	if game.playerIndex(playerID) < 0 {
		return ErrPlayerNotFound
	}

	if game.lastWord.playerID == playerID {
		return ErrCannotVoteOnOwnWord
	}

	if _, voted := game.challenge.votes[playerID]; voted {
		return ErrAlreadyVoted
	}

	return nil
}

// ChallengeableMoverID returns the player whose last word is still open to a
// challenge, if any.
func (game *Game) ChallengeableMoverID() (string, bool) {
	if game.finished || game.challenge != nil || game.lastWord == nil || game.lastWord.settled {
		return "", false
	}

	return game.lastWord.playerID, true
}

// PendingChallenge returns the current tally of the open challenge and
// whether one is pending.
func (game *Game) PendingChallenge() (ChallengeOutcome, bool) {
	if game.challenge == nil {
		return ChallengeOutcome{}, false
	}

	return game.challengeTally(), true
}

func (game *Game) challengeTally() ChallengeOutcome {
	eligible := len(game.players) - 1

	var votesInvalid, votesValid int
	for _, vote := range game.challenge.votes {
		if vote == VoteInvalid {
			votesInvalid++
		} else {
			votesValid++
		}
	}

	return ChallengeOutcome{
		ChallengerID:   game.challenge.challengerID,
		MoverID:        game.lastWord.playerID,
		VotesInvalid:   votesInvalid,
		VotesValid:     votesValid,
		VotesNeeded:    eligible/2 + 1,
		EligibleVoters: eligible,
	}
}

// resolveChallenge settles the challenge as soon as the vote is decided:
// upheld once invalid votes reach a strict majority of eligible voters, or
// rejected once that majority is out of reach.
func (game *Game) resolveChallenge() (ChallengeOutcome, error) {
	outcome := game.challengeTally()

	undecided := outcome.EligibleVoters - outcome.VotesInvalid - outcome.VotesValid

	switch {
	case outcome.VotesInvalid >= outcome.VotesNeeded:
		outcome.Resolved = true
		outcome.Upheld = true

		rescinded, err := game.rescindLastWord()
		if err != nil {
			return ChallengeOutcome{}, fmt.Errorf("rescinding word: %w", err)
		}
		outcome.RescindedWord = &rescinded

		game.challenge = nil
	case outcome.VotesInvalid+undecided < outcome.VotesNeeded:
		outcome.Resolved = true
		game.lastWord.settled = true
		game.challenge = nil
	}

	return outcome, nil
}

// rescindLastWord removes the last played word, returns the letters it drew
// to the pool, and hands the spent letters back to the player. The player's
// turn stays consumed: an upheld challenge forfeits it.
func (game *Game) rescindLastWord() (Word, error) {
	mover := &game.players[game.playerIndex(game.lastWord.playerID)]

	lastTurn := mover.turns[len(mover.turns)-1]
	mover.turns = mover.turns[:len(mover.turns)-1]

	placed := game.board.words
	rescinded := placed[len(placed)-1]

	if err := game.board.removeLastWord(); err != nil {
		return Word{}, fmt.Errorf("removing word from board: %w", err)
	}

	mover.takeLetters(game.pool[game.poolIndex-lastTurn.LettersDrawn : game.poolIndex])
	game.poolIndex -= lastTurn.LettersDrawn
	mover.giveLetters(lettersFromMap(lastTurn.LettersUsed))

	// shuffle so the same letters cannot simply be drawn again
	game.shufflePoolTail()

	game.lastWord = nil

	return rescinded, nil
}

// FindPlacements returns every legal placement of the given letters that
// passes through the given point, ordered by points descending.
func (game *Game) FindPlacements(playerID string, point Point, letters string) ([]PlacementResult, error) {
	if !game.started {
		return nil, ErrGameNotStarted
	}

	var placements []PlacementResult
	for _, direction := range []Direction{DirectionHorizontal, DirectionVertical} {
		for offset := range len(letters) {
			deltaColumn, deltaRow := direction.Vector(-offset)
			word := NewWord(point.Offset(deltaColumn, deltaRow), direction, letters)

			result, err := game.checkWord(playerID, word)
			if err != nil {
				continue
			}

			placements = append(placements, result)
		}
	}

	if len(placements) == 0 {
		return nil, ErrCannotPlayWord
	}

	sort.SliceStable(placements, func(first, second int) bool {
		return placements[first].Points > placements[second].Points
	})

	return placements, nil
}

// checkWord validates the placement against the board and the player's
// rack, substituting blank tiles for letters the player lacks.
func (game *Game) checkWord(playerID string, word Word) (PlacementResult, error) {
	if !game.started {
		return PlacementResult{}, ErrGameNotStarted
	}

	playerIndex := game.playerIndex(playerID)
	if playerIndex < 0 {
		return PlacementResult{}, ErrPlayerNotFound
	}

	result, err := game.board.tryWordPlacement(word)
	if err != nil {
		return PlacementResult{}, fmt.Errorf("checking placement: %w", err)
	}

	canPlay, blanks := game.players[playerIndex].hasLettersWithBlanks(result.LettersUsed)
	if !canPlay {
		return PlacementResult{}, ErrCannotPlayWord
	}

	if len(blanks) > 0 {
		blankPoints := make([]Point, 0, len(blanks))
		for point := range blanks {
			blankPoints = append(blankPoints, point)
		}

		result, err = game.board.tryWordPlacement(word.WithBlanks(blankPoints...))
		if err != nil {
			return PlacementResult{}, fmt.Errorf("checking placement with blanks: %w", err)
		}
	}

	return result, nil
}

func (game *Game) assertTurn(playerID string) error {
	if !game.started {
		return ErrGameNotStarted
	}

	if game.finished {
		return ErrGameFinished
	}

	if game.challenge != nil {
		return ErrChallengePending
	}

	if game.playerIndex(playerID) < 0 {
		return ErrPlayerNotFound
	}

	if game.players[game.turn].id != playerID {
		return ErrNotYourTurn
	}

	return nil
}

func (game *Game) playerIndex(playerID string) int {
	for index := range game.players {
		if game.players[index].id == playerID {
			return index
		}
	}

	return -1
}

// fillPlayerRack tops the player's rack up from the pool and returns how
// many letters were drawn. An exhausted pool is not an error; the rack just
// stays short.
func (game *Game) fillPlayerRack(player *Player) int {
	needed := game.config.RackSize - len(player.letters)
	if needed <= 0 {
		return 0
	}

	needed = min(needed, game.LettersRemaining())
	if needed == 0 {
		return 0
	}

	player.giveLetters(game.pool[game.poolIndex : game.poolIndex+needed])
	game.poolIndex += needed

	return needed
}

func (game *Game) settleLastWord() {
	if game.lastWord != nil {
		game.lastWord.settled = true
	}
}

// endScorelessTurn counts a pass or exchange and either ends the game or
// moves to the next player.
func (game *Game) endScorelessTurn() {
	game.scorelessTurns++
	if game.scorelessTurns >= scorelessRoundLimit*len(game.players) {
		game.finish("")
		return
	}

	game.advanceTurn()
}

func (game *Game) advanceTurn() {
	game.turn++
	if game.turn >= len(game.players) {
		game.turn = 0
		game.round++
	}
}

func (game *Game) shufflePoolTail() {
	tail := game.pool[game.poolIndex:]
	rand.Shuffle(len(tail), func(first, second int) {
		tail[first], tail[second] = tail[second], tail[first]
	})
}

// finish ends the game: every player forfeits the value of the letters left
// on their rack, and a player who went out gains what the others forfeited.
func (game *Game) finish(goingOutPlayerID string) {
	game.finished = true

	var forfeitTotal int
	for index := range game.players {
		player := &game.players[index]
		if player.id == goingOutPlayerID {
			continue
		}

		var rackValue int
		for _, letter := range player.letters {
			rackValue += game.config.LetterPoints[letter]
		}

		player.finalAdjustment = -rackValue
		forfeitTotal += rackValue
	}

	if goingOutIndex := game.playerIndex(goingOutPlayerID); goingOutIndex >= 0 {
		game.players[goingOutIndex].finalAdjustment = forfeitTotal
	}

	game.winnerIDs = game.computeWinnerIDs()
}

func (game *Game) computeWinnerIDs() []string {
	best := math.MinInt

	var winnerIDs []string
	for _, player := range game.players {
		score := player.Score()
		switch {
		case score > best:
			best = score
			winnerIDs = []string{player.id}
		case score == best:
			winnerIDs = append(winnerIDs, player.id)
		}
	}

	return winnerIDs
}

func lettersFromMap(lettersByPoint map[Point]rune) []rune {
	var letters []rune
	for _, letter := range lettersByPoint {
		letters = append(letters, letter)
	}

	return letters
}
