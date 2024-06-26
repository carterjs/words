package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/carterjs/words/internal/game"
	_ "modernc.org/sqlite"
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
	_, err := store.db.Exec(`
		CREATE TABLE IF NOT EXISTS games (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			passphrase_hash BLOB NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	_, err = store.db.Exec(`
		CREATE TABLE IF NOT EXISTS players (
			id TEXT PRIMARY KEY,
			game_id TEXT NOT NULL,
			name TEXT NOT NULL,
			status TEXT NOT NULL,
			letters TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	_, err = store.db.Exec(`
		CREATE TABLE IF NOT EXISTS turns (
			id TEXT PRIMARY KEY,
			game_id TEXT NOT NULL,
			player_id TEXT NOT NULL,
			round INTEGER NOT NULL,
			word_x INTEGER NOT NULL,
			word_y INTEGER NOT NULL,
			word_direction TEXT NOT NULL,
			word_value TEXT NOT NULL,
			points INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func (store *SQLite) SaveGame(ctx context.Context, game game.Game) error {
	_, err := store.db.ExecContext(
		ctx,
		"INSERT INTO games (id, name, passphrase_hash) VALUES (?, ?, ?)",
		game.ID,
		game.Name,
		game.PassphraseHash,
	)
	if err != nil {
		return err
	}

	return nil
}

func (store *SQLite) GetGameByID(ctx context.Context, id string) (*game.Game, error) {
	row := store.db.QueryRowContext(
		ctx,
		"SELECT id, name, passphrase_hash FROM games WHERE id = ?",
		id,
	)
	var g game.Game
	err := row.Scan(&g.ID, &g.Name, &g.PassphraseHash)
	if err != nil {
		return nil, err
	}

	return &g, nil
}

func (store *SQLite) SavePlayer(ctx context.Context, player game.Player) error {
	_, err := store.db.ExecContext(
		ctx,
		"INSERT INTO players (id, game_id, name, status, letters) VALUES (?, ?, ?, ?, ?)",
		player.ID,
		player.GameID,
		player.Name,
		player.Status,
		string(player.Letters),
	)
	if err != nil {
		return err
	}

	return nil
}

func (store *SQLite) GetPlayersByGameID(ctx context.Context, gameID string) ([]game.Player, error) {
	rows, err := store.db.QueryContext(
		ctx,
		"SELECT id, name, status, letters FROM players WHERE game_id = ?",
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
		players = append(players, p)
	}

	return players, nil
}

func (store *SQLite) GetPlayerByID(ctx context.Context, playerID string) (*game.Player, error) {
	row := store.db.QueryRowContext(
		ctx,
		"SELECT id, game_id, name, status, letters FROM players WHERE id = ?",
		playerID,
	)
	var p game.Player
	err := row.Scan(&p.ID, &p.GameID, &p.Name, &p.Status, &p.Letters)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (store *SQLite) SaveTurn(ctx context.Context, turn game.Turn) error {
	_, err := store.db.ExecContext(
		ctx,
		"INSERT INTO turns (id, game_id, player_id, round, word_x, word_y, word_direction, word_value, points) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		turn.ID,
		turn.GameID,
		turn.PlayerID,
		turn.Round,
		turn.Word.X,
		turn.Word.Y,
		turn.Word.Direction,
		turn.Word.Value,
		turn.Points,
	)
	if err != nil {
		return err
	}

	return nil
}

func (store *SQLite) GetTurnsByGameID(ctx context.Context, gameID string) ([]game.Turn, error) {
	rows, err := store.db.QueryContext(
		ctx,
		"SELECT id, player_id, round, word_x, word_y, word_direction, word_value, points FROM turns WHERE game_id = ?",
		gameID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var turns []game.Turn
	for rows.Next() {
		var t game.Turn
		if err := rows.Scan(&t.ID, &t.PlayerID, &t.Round, &t.Word.X, &t.Word.Y, &t.Word.Direction, &t.Word.Value, &t.Points); err != nil {
			return nil, err
		}
		turns = append(turns, t)
	}

	return turns, nil
}
