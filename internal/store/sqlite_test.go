package store_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/carterjs/words/internal/game"
	"github.com/carterjs/words/internal/store"
	"github.com/stretchr/testify/assert"
)

func TestSQLite_Games(t *testing.T) {

	t.Run("set and get game", func(t *testing.T) {
		db := newTestDB(t)

		ctx := context.Background()
		testGame := game.Game{
			ID:              "game1",
			Name:            "game1",
			PassphraseHash:  []byte("passphrase"),
			Round:           1,
			ConfigurationID: "config1",
			Letters:         []rune("abc"),
		}
		err := db.CreateGame(ctx, testGame)
		assert.NoError(t, err)

		gotGame, err := db.GetGameByID(ctx, "game1")
		assert.NoError(t, err)
		assert.Equal(t, testGame, *gotGame)

		testGame.Letters = []rune("def")
		err = db.UpdateGame(ctx, testGame)
		assert.NoError(t, err)

		gotGame, err = db.GetGameByID(ctx, "game1")
		assert.NoError(t, err)
		assert.Equal(t, testGame, *gotGame)
	})

	t.Run("update game that doesn't exist", func(t *testing.T) {
		db := newTestDB(t)

		ctx := context.Background()
		testGame := game.Game{
			ID:              "game1",
			Name:            "game1",
			PassphraseHash:  []byte("passphrase"),
			Round:           1,
			ConfigurationID: "config1",
			Letters:         []rune("abc"),
		}
		err := db.UpdateGame(ctx, testGame)
		assert.ErrorIs(t, err, game.ErrGameNotFound)
	})

	t.Run("id conflict", func(t *testing.T) {
		db := newTestDB(t)

		ctx := context.Background()
		testGame := game.Game{
			ID:              "game1",
			Name:            "game1",
			PassphraseHash:  []byte("passphrase"),
			Round:           1,
			ConfigurationID: "config1",
			Letters:         []rune("abc"),
		}
		err := db.CreateGame(ctx, testGame)
		assert.NoError(t, err)

		err = db.CreateGame(ctx, testGame)
		assert.ErrorIs(t, err, store.ErrPrimaryKeyConflict)
	})

	t.Run("get game that does not exist", func(t *testing.T) {
		db := newTestDB(t)

		ctx := context.Background()
		gotGame, err := db.GetGameByID(ctx, "game1")
		assert.ErrorIs(t, err, game.ErrGameNotFound)
		assert.Nil(t, gotGame)
	})
}

func TestSQLite_Players(t *testing.T) {
	t.Run("set and get player", func(t *testing.T) {
		db := newTestDB(t)

		ctx := context.Background()
		testPlayer := game.Player{
			ID:     "player1",
			GameID: "game1",
			Name:   "Alice",
			Status: game.PlayerStatusActive,
			Letters: []rune{
				'a', 'b', 'c',
			},
		}
		err := db.CreatePlayer(ctx, testPlayer)
		assert.NoError(t, err)

		player, err := db.GetPlayerByID(ctx, "player1")
		assert.NoError(t, err)
		assert.Equal(t, testPlayer, *player)

		testPlayer.Letters = []rune("def")
		err = db.UpdatePlayer(ctx, testPlayer)
		assert.NoError(t, err)
		assert.Equal(t, testPlayer, testPlayer)
	})

	t.Run("update player that doesn't exist", func(t *testing.T) {
		db := newTestDB(t)

		ctx := context.Background()
		testPlayer := game.Player{
			ID:     "player1",
			GameID: "game1",
			Name:   "Alice",
			Status: game.PlayerStatusActive,
			Letters: []rune{
				'a', 'b', 'c',
			},
		}
		err := db.UpdatePlayer(ctx, testPlayer)
		assert.ErrorIs(t, err, game.ErrPlayerNotFound)
	})

	t.Run("id conflict", func(t *testing.T) {
		db := newTestDB(t)

		ctx := context.Background()
		testPlayer := game.Player{
			ID:     "player1",
			GameID: "game1",
			Name:   "Alice",
			Status: game.PlayerStatusActive,
			Letters: []rune{
				'a', 'b', 'c',
			},
		}
		err := db.CreatePlayer(ctx, testPlayer)
		assert.NoError(t, err)

		err = db.CreatePlayer(ctx, testPlayer)
		assert.ErrorIs(t, err, store.ErrPrimaryKeyConflict)
	})

	t.Run("get player that does not exist", func(t *testing.T) {
		db := newTestDB(t)

		ctx := context.Background()
		player, err := db.GetPlayerByID(ctx, "player1")
		assert.ErrorIs(t, err, game.ErrPlayerNotFound)
		assert.Nil(t, player)
	})

	t.Run("get players by game id", func(t *testing.T) {
		db := newTestDB(t)

		ctx := context.Background()
		testPlayer1 := game.Player{
			ID:     "player1",
			GameID: "game1",
			Name:   "Alice",
			Status: game.PlayerStatusActive,
			Letters: []rune{
				'a', 'b', 'c',
			},
		}
		err := db.CreatePlayer(ctx, testPlayer1)
		assert.NoError(t, err)

		testPlayer2 := game.Player{
			ID:     "player2",
			GameID: "game1",
			Name:   "Bob",
			Status: game.PlayerStatusActive,
			Letters: []rune{
				'd', 'e', 'f',
			},
		}
		err = db.CreatePlayer(ctx, testPlayer2)
		assert.NoError(t, err)

		players, err := db.GetPlayersByGameID(ctx, "game1")
		assert.NoError(t, err)
		assert.ElementsMatch(t, []game.Player{testPlayer1, testPlayer2}, players)

	})
}

func TestSQLite_Turns(t *testing.T) {

	t.Run("set and get turn", func(t *testing.T) {
		db := newTestDB(t)

		ctx := context.Background()
		testTurn := game.Turn{
			ID:       "turn1",
			GameID:   "game1",
			PlayerID: "player1",
			Word: game.Word{
				X:         0,
				Y:         0,
				Direction: game.DirectionHorizontal,
				Letters:   []rune("hello"),
			},
		}
		err := db.CreateTurn(ctx, testTurn)
		assert.NoError(t, err)

		turn, err := db.GetTurnByID(ctx, "turn1")
		assert.NoError(t, err)
		assert.Equal(t, testTurn, *turn)
	})

	t.Run("get turns for game and round", func(t *testing.T) {
		db := newTestDB(t)

		ctx := context.Background()
		turn1 := game.Turn{
			ID:       "turn1",
			GameID:   "game1",
			Round:    1,
			PlayerID: "player1",
			Word: game.Word{
				X:         0,
				Y:         0,
				Direction: game.DirectionHorizontal,
				Letters:   []rune("hello"),
			},
			Status: game.TurnStatusPlayed,
		}
		err := db.CreateTurn(ctx, turn1)
		assert.NoError(t, err)

		turn2 := game.Turn{
			ID:       "turn2",
			GameID:   "game1",
			Round:    1,
			PlayerID: "player2",
			Word: game.Word{
				X:         0,
				Y:         0,
				Direction: game.DirectionHorizontal,
				Letters:   []rune("world"),
			},
			Status: game.TurnStatusPending,
		}
		err = db.CreateTurn(ctx, turn2)
		assert.NoError(t, err)

		turns, err := db.GetTurnsByGameID(ctx, "game1")
		assert.NoError(t, err)
		assert.ElementsMatch(t, []game.Turn{
			turn1,
			turn2,
		}, turns)

		turns, err = db.GetTurnsByRound(ctx, "game1", 1)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []game.Turn{
			turn1,
			turn2,
		}, turns)
	})

	t.Run("no turns for game", func(t *testing.T) {
		db := newTestDB(t)

		ctx := context.Background()
		turns, err := db.GetTurnsByGameID(ctx, "game1")
		assert.NoError(t, err)
		assert.Empty(t, turns)
	})
}

func TestSQLite_TurnVotes(t *testing.T) {
	db := newTestDB(t)

	err := db.CreateTurnVote(context.Background(), game.TurnVote{
		TurnID:   "turn1",
		PlayerID: "player1",
		Value:    game.TurnVoteValueApprove,
	})
	assert.NoError(t, err)

	votes, err := db.GetTurnVotes(context.Background(), "turn1")
	assert.NoError(t, err)
	assert.ElementsMatch(t, []game.TurnVote{
		{
			TurnID:   "turn1",
			PlayerID: "player1",
			Value:    game.TurnVoteValueApprove,
		},
	}, votes)
}

func TestSQLite_Configurations(t *testing.T) {
	t.Run("set and get configuration", func(t *testing.T) {
		db := newTestDB(t)

		ctx := context.Background()
		testConfiguration := game.Configuration{
			ID:          "config1",
			OwnerID:     "player1",
			Title:       "Configuration 1",
			Description: "This is a test configuration",
			Letters: []game.LetterConfiguration{
				{
					Letter: 'a',
					Count:  1,
					Points: 1,
				},
				{
					Letter: 'b',
					Count:  2,
					Points: 2,
				},
				{
					Letter: 'c',
					Count:  3,
					Points: 3,
				},
			},
		}
		err := db.CreateConfiguration(ctx, testConfiguration)
		assert.NoError(t, err)

		configuration, err := db.GetConfigurationByID(ctx, "config1")
		assert.NoError(t, err)
		assert.Equal(t, testConfiguration, *configuration)
	})

	t.Run("get configuration that does not exist", func(t *testing.T) {
		db := newTestDB(t)

		ctx := context.Background()
		configuration, err := db.GetConfigurationByID(ctx, "config1")
		assert.ErrorIs(t, err, game.ErrConfigurationNotFound)
		assert.Nil(t, configuration)
	})

	t.Run("id conflict", func(t *testing.T) {
		db := newTestDB(t)

		ctx := context.Background()
		testConfiguration := game.Configuration{
			ID:          "config1",
			OwnerID:     "player1",
			Title:       "Configuration 1",
			Description: "This is a test configuration",
			Letters: []game.LetterConfiguration{
				{
					Letter: 'a',
					Count:  1,
					Points: 1,
				},
				{
					Letter: 'b',
					Count:  2,
					Points: 2,
				},
				{
					Letter: 'c',
					Count:  3,
					Points: 3,
				},
			},
		}
		err := db.CreateConfiguration(ctx, testConfiguration)
		assert.NoError(t, err)

		err = db.CreateConfiguration(ctx, testConfiguration)
		assert.ErrorIs(t, err, store.ErrPrimaryKeyConflict)
	})
}
func newTestDB(t *testing.T) *store.SQLite {
	f := filepath.Join(t.TempDir(), "test.db")

	db, err := store.NewSQLite(f)
	assert.NoError(t, err)

	return db
}
