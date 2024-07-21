package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/carterjs/words/internal/game"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

var (
	ErrPrimaryKeyConflict = errors.New("primary key conflict")
)

type SQLite struct {
	db *sql.DB
}

func NewSQLite(filename string) (*SQLite, error) {
	db, err := sql.Open("sqlite", filename)
	if err != nil {
		return nil, err
	}

	store := &SQLite{db: db}

	if err := store.init(); err != nil {
		return nil, fmt.Errorf("failed to initialize store: %w", err)
	}

	return store, nil
}

func (store *SQLite) Close() error {
	return store.db.Close()
}

func (store *SQLite) init() error {
	var errs error
	_, err := store.db.Exec(`
		CREATE TABLE IF NOT EXISTS game (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			passphrase_hash BLOB NOT NULL,
			round INTEGER NOT NULL DEFAULT 1,
			configuration_id TEXT NOT NULL,
			letters TEXT NOT NULL
		)
	`)
	errs = errors.Join(errs, err)

	_, err = store.db.Exec(`
		CREATE TABLE IF NOT EXISTS configuration (
			id TEXT PRIMARY KEY,
			owner_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT NOT NULL
		)
	`)
	errs = errors.Join(errs, err)

	_, err = store.db.Exec(`
		CREATE TABLE IF NOT EXISTS letter_configuration (
			configuration_id TEXT NOT NULL,
			letter TEXT NOT NULL,
			points INTEGER NOT NULL,
			count INTEGER NOT NULL,
			PRIMARY KEY (configuration_id, letter)
		)
		`)
	errs = errors.Join(errs, err)

	_, err = store.db.Exec(`
		CREATE TABLE IF NOT EXISTS player (
			id TEXT PRIMARY KEY,
			game_id TEXT NOT NULL,
			name TEXT NOT NULL,
			status TEXT NOT NULL,
			letters TEXT NOT NULL
		)
	`)
	errs = errors.Join(errs, err)

	_, err = store.db.Exec(`
		CREATE TABLE IF NOT EXISTS turn (
			id TEXT PRIMARY KEY,
			game_id TEXT NOT NULL,
			player_id TEXT NOT NULL,
			round INTEGER NOT NULL,
			word_x INTEGER NOT NULL,
			word_y INTEGER NOT NULL,
			word_direction TEXT NOT NULL,
			word_letters TEXT NOT NULL,
			status TEXT NOT NULL
		)
	`)
	errs = errors.Join(errs, err)

	_, err = store.db.Exec(`
		CREATE TABLE IF NOT EXISTS turn_vote (
			turn_id TEXT NOT NULL,
			player_id TEXT NOT NULL,
			value TEXT NOT NULL,
			PRIMARY KEY (turn_id, player_id)
		)
	`)
	errs = errors.Join(errs, err)

	// TODO: add timestamp
	_, err = store.db.Exec(`
		CREATE TABLE IF NOT EXISTS word_usage (
			word TEXT PRIMARY KEY,
			approvals INTEGER NOT NULL DEFAULT 0,
			rejections INTEGER NOT NULL DEFAULT 0
		)
	`)
	errs = errors.Join(errs, err)

	return errs
}

func (store *SQLite) CreateGame(ctx context.Context, game game.Game) error {
	_, err := store.db.ExecContext(
		ctx,
		"INSERT INTO game (id, name, passphrase_hash, round, configuration_id, letters) VALUES (?, ?, ?, ?, ?, ?)",
		game.ID,
		game.Name,
		game.PassphraseHash,
		game.Round,
		game.ConfigurationID,
		string(game.Letters),
	)
	if err != nil {
		if isConstraintError(err) {
			return ErrPrimaryKeyConflict
		}

		return err
	}

	return nil
}

func isConstraintError(err error) bool {
	if sqliteErr, ok := err.(*sqlite.Error); ok {
		slog.Info("sqlite error", "code", sqliteErr.Code())
		return sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY
	}

	return false
}

func (store *SQLite) UpdateGame(ctx context.Context, newGame game.Game) error {
	result, err := store.db.ExecContext(
		ctx,
		"UPDATE game SET name = ?, passphrase_hash = ?, round = ?, configuration_id = ?, letters = ? WHERE id = ?",
		newGame.Name,
		newGame.PassphraseHash,
		newGame.Round,
		newGame.ConfigurationID,
		string(newGame.Letters),
		newGame.ID,
	)
	if err != nil {
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return game.ErrGameNotFound
	}

	return nil
}

func (store *SQLite) GetGameByID(ctx context.Context, id string) (*game.Game, error) {
	row := store.db.QueryRowContext(
		ctx,
		"SELECT id, name, passphrase_hash, round, configuration_id, letters FROM game WHERE id = ?",
		id,
	)
	var g game.Game
	var letters string
	err := row.Scan(&g.ID, &g.Name, &g.PassphraseHash, &g.Round, &g.ConfigurationID, &letters)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, game.ErrGameNotFound
		}

		return nil, err
	}

	g.Letters = []rune(letters)

	return &g, nil
}

func (store *SQLite) CreatePlayer(ctx context.Context, player game.Player) error {
	_, err := store.db.ExecContext(
		ctx,
		"INSERT INTO player (id, game_id, name, status, letters) VALUES (?, ?, ?, ?, ?)",
		player.ID,
		player.GameID,
		player.Name,
		player.Status,
		string(player.Letters),
	)
	if err != nil {
		if isConstraintError(err) {
			return ErrPrimaryKeyConflict
		}

		return err
	}

	return nil
}

func (store *SQLite) UpdatePlayer(ctx context.Context, player game.Player) error {
	result, err := store.db.ExecContext(
		ctx,
		"UPDATE player SET game_id = ?, name = ?, status = ?, letters = ? WHERE id = ?",
		player.GameID,
		player.Name,
		player.Status,
		string(player.Letters),
		player.ID,
	)
	if err != nil {
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return game.ErrPlayerNotFound
	}

	return nil
}

func (store *SQLite) GetPlayersByGameID(ctx context.Context, gameID string) ([]game.Player, error) {
	rows, err := store.db.QueryContext(
		ctx,
		"SELECT id, name, status, letters FROM player WHERE game_id = ?",
		gameID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []game.Player
	for rows.Next() {
		var p game.Player
		var letters string
		if err := rows.Scan(&p.ID, &p.Name, &p.Status, &letters); err != nil {
			return nil, err
		}
		p.Letters = []rune(letters)
		p.GameID = gameID
		players = append(players, p)
	}

	return players, nil
}

func (store *SQLite) GetPlayerByID(ctx context.Context, playerID string) (*game.Player, error) {
	row := store.db.QueryRowContext(
		ctx,
		"SELECT id, game_id, name, status, letters FROM player WHERE id = ?",
		playerID,
	)
	var p game.Player
	var letters string
	err := row.Scan(&p.ID, &p.GameID, &p.Name, &p.Status, &letters)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, game.ErrPlayerNotFound
		}

		return nil, err
	}
	p.Letters = []rune(letters)

	return &p, nil
}

func (store *SQLite) CreateTurn(ctx context.Context, turn game.Turn) error {
	_, err := store.db.ExecContext(
		ctx,
		"INSERT INTO turn (id, game_id, player_id, round, word_x, word_y, word_direction, word_letters, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		turn.ID,
		turn.GameID,
		turn.PlayerID,
		turn.Round,
		turn.Word.X,
		turn.Word.Y,
		turn.Word.Direction,
		string(turn.Word.Letters),
		turn.Status,
	)
	if err != nil {
		return err
	}

	return nil
}

func (store *SQLite) UpdateTurn(ctx context.Context, turn game.Turn) error {
	_, err := store.db.ExecContext(
		ctx,
		"UPDATE turn SET game_id = ?, player_id = ?, round = ?, word_x = ?, word_y = ?, word_direction = ?, word_letters = ?, status = ? WHERE id = ?",
		turn.GameID,
		turn.PlayerID,
		turn.Round,
		turn.Word.X,
		turn.Word.Y,
		turn.Word.Direction,
		string(turn.Word.Letters),
		turn.Status,
		turn.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (store *SQLite) GetTurnByID(ctx context.Context, id string) (*game.Turn, error) {
	row := store.db.QueryRowContext(
		ctx,
		"SELECT game_id, player_id, round, word_x, word_y, word_direction, word_letters, status FROM turn WHERE id = ?",
		id,
	)
	var t game.Turn
	var letters string
	err := row.Scan(&t.GameID, &t.PlayerID, &t.Round, &t.Word.X, &t.Word.Y, &t.Word.Direction, &letters, &t.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, game.ErrTurnNotFound
		}
	}

	t.ID = id
	t.Word.Letters = []rune(letters)

	return &t, nil
}

func (store *SQLite) GetTurnsByGameID(ctx context.Context, gameID string) ([]game.Turn, error) {
	rows, err := store.db.QueryContext(
		ctx,
		"SELECT id, player_id, round, word_x, word_y, word_direction, word_letters, status FROM turn WHERE game_id = ?",
		gameID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var turns []game.Turn
	for rows.Next() {
		var t game.Turn
		var letters string
		if err := rows.Scan(&t.ID, &t.PlayerID, &t.Round, &t.Word.X, &t.Word.Y, &t.Word.Direction, &letters, &t.Status); err != nil {
			return nil, err
		}
		t.GameID = gameID
		t.Word.Letters = []rune(letters)
		turns = append(turns, t)
	}

	return turns, nil
}

func (store *SQLite) GetTurnsByRound(ctx context.Context, gameID string, round int) ([]game.Turn, error) {
	rows, err := store.db.QueryContext(
		ctx,
		"SELECT id, player_id, word_x, word_y, word_direction, word_letters, status FROM turn WHERE game_id = ? AND round = ?",
		gameID,
		round,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var turns []game.Turn
	for rows.Next() {
		var t game.Turn
		var letters string
		if err := rows.Scan(&t.ID, &t.PlayerID, &t.Word.X, &t.Word.Y, &t.Word.Direction, &letters, &t.Status); err != nil {
			return nil, err
		}
		t.GameID = gameID
		t.Round = round
		t.Word.Letters = []rune(letters)
		turns = append(turns, t)
	}

	return turns, nil
}

func (store *SQLite) CreateTurnVote(ctx context.Context, vote game.TurnVote) error {
	_, err := store.db.ExecContext(
		ctx,
		"INSERT INTO turn_vote (turn_id, player_id, value) VALUES (?, ?, ?)",
		vote.TurnID,
		vote.PlayerID,
		vote.Value,
	)
	if err != nil {
		return err
	}

	return nil
}

func (store *SQLite) GetTurnVotes(ctx context.Context, turnID string) ([]game.TurnVote, error) {
	rows, err := store.db.QueryContext(
		ctx,
		"SELECT player_id, value FROM turn_vote WHERE turn_id = ?",
		turnID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var votes []game.TurnVote
	for rows.Next() {
		var v game.TurnVote
		if err := rows.Scan(&v.PlayerID, &v.Value); err != nil {
			return nil, err
		}
		v.TurnID = turnID
		votes = append(votes, v)
	}

	return votes, nil
}

func (store *SQLite) CreateConfiguration(ctx context.Context, configuration game.Configuration) error {
	_, err := store.db.ExecContext(
		ctx,
		"INSERT INTO configuration (id, owner_id, title, description) VALUES (?, ?, ?, ?)",
		configuration.ID,
		configuration.OwnerID,
		configuration.Title,
		configuration.Description,
	)
	if err != nil {
		if isConstraintError(err) {
			return ErrPrimaryKeyConflict
		}

		return err
	}

	for _, letterConfig := range configuration.Letters {
		_, err := store.db.ExecContext(
			ctx,
			"INSERT INTO letter_configuration (configuration_id, letter, points, count) VALUES (?, ?, ?, ?)",
			configuration.ID,
			string(letterConfig.Letter),
			letterConfig.Points,
			letterConfig.Count,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (store *SQLite) GetConfigurationByID(ctx context.Context, id string) (*game.Configuration, error) {
	row := store.db.QueryRowContext(
		ctx,
		"SELECT id, owner_id, title, description FROM configuration WHERE id = ?",
		id,
	)
	var c game.Configuration
	err := row.Scan(&c.ID, &c.OwnerID, &c.Title, &c.Description)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, game.ErrConfigurationNotFound
		}

		return nil, err
	}

	rows, err := store.db.QueryContext(
		ctx,
		"SELECT letter, points, count FROM letter_configuration WHERE configuration_id = ?",
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var lc game.LetterConfiguration
		var letter string
		if err := rows.Scan(&letter, &lc.Points, &lc.Count); err != nil {
			return nil, err
		}
		lc.Letter = []rune(letter)[0]
		c.Letters = append(c.Letters, lc)
	}

	return &c, nil
}

func (store *SQLite) GetWordStats(ctx context.Context, word string) (game.WordStats, error) {
	row := store.db.QueryRowContext(
		ctx,
		"SELECT count(word), coalesce(sum(approvals), 0), coalesce(sum(rejections), 0) FROM word_usage WHERE word = ?",
		word,
	)
	var stats game.WordStats
	err := row.Scan(&stats.Usages, &stats.Approvals, &stats.Rejections)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return game.WordStats{}, nil
		}

		return game.WordStats{}, err
	}

	return stats, nil
}

func (store *SQLite) SaveWordUsage(ctx context.Context, word string, approvals, rejections int) error {
	_, err := store.db.ExecContext(
		ctx,
		"INSERT INTO word_usage (word, approvals, rejections) VALUES (?, ?, ?)",
		word,
		approvals,
		rejections,
	)
	if err != nil {
		return err
	}

	slog.Info("saving word usage!", "word", word, "approvals", approvals, "rejections", rejections)

	return nil
}
