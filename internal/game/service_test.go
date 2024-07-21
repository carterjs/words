package game_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/carterjs/words/internal/game"
	"github.com/carterjs/words/internal/store"
	"github.com/stretchr/testify/assert"
)

func TestGameService(t *testing.T) {
	validWords := []game.Word{
		{
			X:         0,
			Y:         0,
			Direction: game.DirectionHorizontal,
			Letters:   []rune("hello"),
		},
		{
			X:         0,
			Y:         0,
			Direction: game.DirectionVertical,
			Letters:   []rune("hola"),
		},
		{
			X:         0,
			Y:         3,
			Direction: game.DirectionHorizontal,
			Letters:   []rune("allowance"),
		},
	}

	t.Run("one player, one turn, one vote", func(t *testing.T) {
		gameService := newTestGameService(t)
		testGame, players := initializeGameWithPlayers(t, gameService, "a")
		onlyPlayer := players[0]

		submitTurn(t, gameService, testGame.ID, onlyPlayer.ID, validWords[0])
		assertRoundMatches(t, gameService, testGame.ID, 2)
	})

	t.Run("two players, two turns, two votes", func(t *testing.T) {
		gameService := newTestGameService(t)
		testGame, players := initializeGameWithPlayers(t, gameService, "a", "b")
		playerA := players[0]
		playerB := players[1]

		playerATurn := submitTurn(t, gameService, testGame.ID, playerA.ID, validWords[0])
		playerBTurn := submitTurn(t, gameService, testGame.ID, playerB.ID, validWords[1])

		voteOnTurn(t, gameService, testGame.ID, playerATurn.ID, playerB.ID, true)
		assertRoundMatches(t, gameService, testGame.ID, 1)

		voteOnTurn(t, gameService, testGame.ID, playerBTurn.ID, playerA.ID, true)
		assertRoundMatches(t, gameService, testGame.ID, 2)
	})

	t.Run("three players, three turns, three votes", func(t *testing.T) {
		gameService := newTestGameService(t)
		testGame, players := initializeGameWithPlayers(t, gameService, "a", "b", "c")
		playerA := players[0]
		playerB := players[1]
		playerC := players[2]

		playerATurn := submitTurn(t, gameService, testGame.ID, playerA.ID, validWords[0])
		playerBTurn := submitTurn(t, gameService, testGame.ID, playerB.ID, validWords[1])
		playerCTurn := submitTurn(t, gameService, testGame.ID, playerC.ID, validWords[2])

		// both other players vote for player A's turn
		voteOnTurn(t, gameService, testGame.ID, playerATurn.ID, playerB.ID, true)
		voteOnTurn(t, gameService, testGame.ID, playerATurn.ID, playerC.ID, true)
		assertRoundMatches(t, gameService, testGame.ID, 1)

		// both other players vote for player B's turn
		voteOnTurn(t, gameService, testGame.ID, playerBTurn.ID, playerA.ID, true)
		voteOnTurn(t, gameService, testGame.ID, playerBTurn.ID, playerC.ID, true)
		assertRoundMatches(t, gameService, testGame.ID, 1)

		// both other players vote for player C's turn and the round advances
		voteOnTurn(t, gameService, testGame.ID, playerCTurn.ID, playerA.ID, true)
		assertRoundMatches(t, gameService, testGame.ID, 1)
		voteOnTurn(t, gameService, testGame.ID, playerCTurn.ID, playerB.ID, true)
		assertRoundMatches(t, gameService, testGame.ID, 2)

	})
}

func newTestGameService(t *testing.T) *game.Service {
	testLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	store, err := store.NewSQLite(filepath.Join(t.TempDir(), "test.db"))
	assert.NoError(t, err)

	return game.NewService(testLogger, store)
}

func initializeGameWithPlayers(t *testing.T, gameService *game.Service, names ...string) (*game.Game, []game.Player) {
	t.Helper()

	ctx := context.Background()

	passphrase := "passphrase"
	newGame, err := gameService.CreateGame(ctx, "game1", passphrase)
	assert.NoError(t, err)

	var players []game.Player
	for _, name := range names {
		player, err := gameService.AddPlayerToGame(ctx, newGame.ID, name, passphrase)
		assert.NoError(t, err)

		players = append(players, *player)
	}

	return newGame, players
}

func submitTurn(t *testing.T, gameService *game.Service, gameID, playerID string, word game.Word) *game.Turn {
	t.Helper()

	ctx := context.Background()
	turn, err := gameService.SubmitTurn(ctx, gameID, playerID, word)
	assert.NoError(t, err)

	return turn
}

func voteOnTurn(t *testing.T, gameService *game.Service, gameID, turnID, voterID string, vote bool) {
	t.Helper()

	ctx := context.Background()
	err := gameService.VoteOnTurn(ctx, gameID, turnID, voterID, vote)
	assert.NoError(t, err)
}

func assertRoundMatches(t *testing.T, gameService *game.Service, gameID string, round int) {
	t.Helper()

	ctx := context.Background()
	game, err := gameService.GetGameByID(ctx, gameID)
	assert.NoError(t, err)
	assert.Equal(t, round, game.Round)
}
