package words_test

import (
	"fmt"
	"testing"

	"github.com/carterjs/words/internal/words"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGame_Start(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		players    int
		startTwice bool
		wantErr    error
	}{
		{name: "rejects a game with no players", players: 0, wantErr: words.ErrNotEnoughPlayers},
		{name: "rejects starting twice", players: 1, startTwice: true, wantErr: words.ErrGameStarted},
		{name: "deals every player a full rack", players: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			game := newLobbyGame(t, test.players, testConfig(map[rune]int{'A': 20}, 3))

			err := game.Start()
			if test.startTwice {
				err = game.Start()
			}

			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			for _, player := range game.Players() {
				assert.Len(t, player.Letters(), 3)
			}
		})
	}
}

func TestGame_PlayWord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		players      int
		config       words.Config
		skipStart    bool
		prePlays     []words.Word
		outOfTurn    bool
		word         words.Word
		wantErr      error
		wantPoints   int
		wantRackLen  int
		wantFinished bool
	}{
		{name: "rejects an unstarted game", players: 1, config: testConfig(map[rune]int{'A': 20}, 3), skipStart: true, word: horizontal(0, 0, "AA"), wantErr: words.ErrGameNotStarted},
		{name: "rejects playing out of turn", players: 2, config: testConfig(map[rune]int{'A': 20}, 3), outOfTurn: true, word: horizontal(0, 0, "AA"), wantErr: words.ErrNotYourTurn},
		{name: "rejects letters the player lacks", players: 1, config: testConfig(map[rune]int{'A': 20}, 3), word: horizontal(0, 0, "BB"), wantErr: words.ErrCannotPlayWord},
		{name: "scores the word and refills the rack", players: 1, config: testConfig(map[rune]int{'A': 20}, 3), word: horizontal(0, 0, "AA"), wantPoints: 2, wantRackLen: 3},
		{name: "substitutes blanks for missing letters", players: 1, config: testConfig(map[rune]int{words.BlankLetter: 3}, 3), word: horizontal(0, 0, "AB"), wantPoints: 0, wantRackLen: 1},
		{name: "finishes when the pool and rack empty", players: 1, config: testConfig(map[rune]int{'A': 4}, 2), prePlays: []words.Word{horizontal(0, 0, "AA"), vertical(0, 0, "AA")}, word: vertical(1, 0, "AA"), wantPoints: 4, wantRackLen: 0, wantFinished: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			game := newLobbyGame(t, test.players, test.config)
			if !test.skipStart {
				require.NoError(t, game.Start())
			}

			for _, word := range test.prePlays {
				playCurrent(t, game, word)
			}

			playerID := game.Players()[0].ID()
			if test.outOfTurn {
				playerID = game.Players()[1].ID()
			}

			result, err := game.PlayWord(playerID, test.word)

			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.wantPoints, result.Points)
			assert.Equal(t, test.wantFinished, game.Finished())
			assert.Len(t, mustPlayer(t, game, playerID).Letters(), test.wantRackLen)
		})
	}
}

func TestGame_PassTurn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		passes       int
		wantFinished bool
		wantScore    int
	}{
		{name: "advances to the next player", passes: 1},
		{name: "ends the game after two scoreless rounds", passes: 4, wantFinished: true, wantScore: -3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			game := newStartedGame(t, 2, testConfig(map[rune]int{'A': 20}, 3))
			first := game.Players()[0]

			for range test.passes {
				require.NoError(t, game.PassTurn(game.CurrentPlayerID()))
			}

			assert.Equal(t, test.wantFinished, game.Finished())
			assert.Equal(t, test.wantScore, mustPlayer(t, game, first.ID()).Score())
		})
	}
}

func TestGame_ExchangeLetters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		distribution  map[rune]int
		letters       string
		wantErr       error
		wantRemaining int
	}{
		{name: "swaps letters and consumes the turn", distribution: map[rune]int{'A': 10}, letters: "A", wantRemaining: 4},
		{name: "rejects letters the player lacks", distribution: map[rune]int{'A': 10}, letters: "Z", wantErr: words.ErrMissingLetters},
		{name: "rejects exchanges larger than the pool", distribution: map[rune]int{'A': 7}, letters: "AA", wantErr: words.ErrNotEnoughLettersInPool},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			game := newStartedGame(t, 2, testConfig(test.distribution, 3))
			playerID := game.CurrentPlayerID()

			err := game.ExchangeLetters(playerID, []rune(test.letters))

			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Len(t, mustPlayer(t, game, playerID).Letters(), 3)
			assert.Equal(t, test.wantRemaining, game.LettersRemaining())
			assert.NotEqual(t, playerID, game.CurrentPlayerID())
		})
	}
}

func TestGame_Challenge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		players      int
		skipPlay     bool
		ownWord      bool
		wantErr      error
		wantResolved bool
		wantUpheld   bool
	}{
		{name: "auto-upholds with two players", players: 2, wantResolved: true, wantUpheld: true},
		{name: "stays open with three players", players: 3},
		{name: "rejects challenging your own word", players: 2, ownWord: true, wantErr: words.ErrCannotChallengeOwnWord},
		{name: "rejects a challenge with no word", players: 2, skipPlay: true, wantErr: words.ErrNothingToChallenge},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			game := newStartedGame(t, test.players, testConfig(map[rune]int{'A': 30}, 3))
			moverID := game.CurrentPlayerID()
			if !test.skipPlay {
				playCurrent(t, game, horizontal(0, 0, "AA"))
			}

			challengerID := game.Players()[1].ID()
			if test.ownWord {
				challengerID = moverID
			}

			outcome, err := game.Challenge(challengerID)

			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.wantResolved, outcome.Resolved)
			assert.Equal(t, test.wantUpheld, outcome.Upheld)
			assert.Equal(t, test.wantUpheld, len(game.Board().Words()) == 0)

			if !test.wantResolved {
				_, err := game.PlayWord(game.CurrentPlayerID(), horizontal(0, 1, "AA"))
				assert.ErrorIs(t, err, words.ErrChallengePending)
			}
		})
	}
}

func TestGame_CastVote(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		skipChallenge bool
		voterIndex    int
		vote          words.Vote
		wantErr       error
		wantUpheld    bool
	}{
		{name: "upholds the challenge on majority invalid", voterIndex: 2, vote: words.VoteInvalid, wantUpheld: true},
		{name: "settles the word when the vote splits", voterIndex: 2, vote: words.VoteValid},
		{name: "rejects the mover voting", voterIndex: 0, vote: words.VoteValid, wantErr: words.ErrCannotVoteOnOwnWord},
		{name: "rejects a duplicate vote", voterIndex: 1, vote: words.VoteInvalid, wantErr: words.ErrAlreadyVoted},
		{name: "rejects an unrecognized vote", voterIndex: 2, vote: words.Vote("MAYBE"), wantErr: words.ErrInvalidVote},
		{name: "rejects a vote with no challenge", skipChallenge: true, voterIndex: 1, vote: words.VoteValid, wantErr: words.ErrNoPendingChallenge},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			game := newStartedGame(t, 3, testConfig(map[rune]int{'A': 30}, 3))
			playCurrent(t, game, horizontal(0, 0, "AA"))

			if !test.skipChallenge {
				_, err := game.Challenge(game.Players()[1].ID())
				require.NoError(t, err)
			}

			voterID := game.Players()[test.voterIndex].ID()
			outcome, err := game.CastVote(voterID, test.vote)

			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			assert.True(t, outcome.Resolved)
			assert.Equal(t, test.wantUpheld, outcome.Upheld)
			assert.Equal(t, test.wantUpheld, len(game.Board().Words()) == 0)

			_, err = game.Challenge(game.Players()[1].ID())
			assert.ErrorIs(t, err, words.ErrNothingToChallenge)
		})
	}
}

func TestGame_FindPlacements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		skipStart      bool
		setup          bool
		letters        string
		wantErr        error
		wantPlacements int
	}{
		{name: "rejects an unstarted game", skipStart: true, letters: "AA", wantErr: words.ErrGameNotStarted},
		{name: "rejects letters with no placement", letters: "ZZ", wantErr: words.ErrCannotPlayWord},
		{name: "finds placements through the point", letters: "AA", wantPlacements: 4},
		{name: "fills placeholders from board letters", setup: true, letters: "*A", wantPlacements: 1},
		{name: "rejects placeholders over empty cells", letters: "*A", wantErr: words.ErrCannotPlayWord},
		{name: "rejects a word of only placeholders", setup: true, letters: "*", wantErr: words.ErrCannotPlayWord},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			game := newLobbyGame(t, 1, testConfig(map[rune]int{'A': 20}, 3))
			if !test.skipStart {
				require.NoError(t, game.Start())
			}
			if test.setup {
				_, err := game.PlayWord(game.Players()[0].ID(), horizontal(0, 0, "AA"))
				require.NoError(t, err)
			}

			placements, err := game.FindPlacements(game.Players()[0].ID(), words.NewPoint(0, 0), test.letters)

			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Len(t, placements, test.wantPlacements)
		})
	}
}

func TestNewGameFromState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "roundtrips a game with an open challenge"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			game := newStartedGame(t, 3, testConfig(map[rune]int{'A': 30}, 3))
			playCurrent(t, game, horizontal(0, 0, "AA"))
			_, err := game.Challenge(game.Players()[1].ID())
			require.NoError(t, err)

			rebuilt, err := words.NewGameFromState(game.State())
			require.NoError(t, err)
			assert.Equal(t, game.State(), rebuilt.State())

			// the rebuilt game keeps playing: the third player settles the vote
			outcome, err := rebuilt.CastVote(rebuilt.Players()[2].ID(), words.VoteInvalid)
			require.NoError(t, err)
			assert.True(t, outcome.Upheld)
		})
	}
}

func newLobbyGame(t *testing.T, playerCount int, config words.Config) *words.Game {
	t.Helper()

	game := words.NewGame(config)
	for index := range playerCount {
		_, err := game.AddPlayer(fmt.Sprintf("player-%d", index))
		require.NoError(t, err)
	}

	return game
}

func newStartedGame(t *testing.T, playerCount int, config words.Config) *words.Game {
	t.Helper()

	game := newLobbyGame(t, playerCount, config)
	require.NoError(t, game.Start())

	return game
}

func playCurrent(t *testing.T, game *words.Game, word words.Word) words.PlacementResult {
	t.Helper()

	result, err := game.PlayWord(game.CurrentPlayerID(), word)
	require.NoError(t, err)

	return result
}

func mustPlayer(t *testing.T, game *words.Game, playerID string) words.Player {
	t.Helper()

	player, exists := game.PlayerByID(playerID)
	require.True(t, exists)

	return player
}

func testConfig(distribution map[rune]int, rackSize int) words.Config {
	return words.Config{
		LetterDistribution: distribution,
		LetterPoints:       map[rune]int{'A': 1, 'B': 2, words.BlankLetter: 0},
		RackSize:           rackSize,
	}
}

func horizontal(column, row int, letters string) words.Word {
	return words.NewWord(words.NewPoint(column, row), words.DirectionHorizontal, letters)
}

func vertical(column, row int, letters string) words.Word {
	return words.NewWord(words.NewPoint(column, row), words.DirectionVertical, letters)
}
