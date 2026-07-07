// Package store persists games to the local filesystem as gzipped JSON
// snapshots, satisfying the words service's Store contract.
package store

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/carterjs/words/internal/words"
)

// directoryPermissions is the mode for the games directory.
const directoryPermissions = 0o755

// FS stores each game as a gzipped JSON snapshot in a directory.
type FS struct {
	directory string
}

// NewFS returns a store writing games to the given directory.
func NewFS(directory string) *FS {
	return &FS{
		directory: directory,
	}
}

// SaveGame writes the game's snapshot to disk, replacing any previous one.
func (fileStore *FS) SaveGame(ctx context.Context, game *words.Game) error {
	if err := os.MkdirAll(fileStore.directory, directoryPermissions); err != nil {
		return fmt.Errorf("creating games directory: %w", err)
	}

	file, err := os.Create(fileStore.gameFile(game.ID()))
	if err != nil {
		return fmt.Errorf("creating game file: %w", err)
	}
	defer file.Close()

	compressor := gzip.NewWriter(file)
	if err := json.NewEncoder(compressor).Encode(game.State()); err != nil {
		return fmt.Errorf("encoding game: %w", err)
	}

	if err := compressor.Close(); err != nil {
		return fmt.Errorf("flushing game: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("closing game file: %w", err)
	}

	return nil
}

// GameByID reads a game's snapshot from disk and rebuilds it. A missing file
// is reported as words.ErrGameNotFound.
func (fileStore *FS) GameByID(ctx context.Context, gameID string) (*words.Game, error) {
	file, err := os.Open(fileStore.gameFile(gameID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, words.ErrGameNotFound
		}

		return nil, fmt.Errorf("opening game file: %w", err)
	}
	defer file.Close()

	decompressor, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("decompressing game: %w", err)
	}
	defer decompressor.Close()

	var state words.GameState
	if err := json.NewDecoder(decompressor).Decode(&state); err != nil {
		return nil, fmt.Errorf("decoding game: %w", err)
	}

	game, err := words.NewGameFromState(state)
	if err != nil {
		return nil, fmt.Errorf("rebuilding game: %w", err)
	}

	return game, nil
}

func (fileStore *FS) gameFile(gameID string) string {
	return filepath.Join(fileStore.directory, gameID+".json.gz")
}
