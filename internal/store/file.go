package store

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/carterjs/words/internal/words"
	"os"
	"path/filepath"
)

type FS struct {
	dir string
}

func NewFS(dir string) *FS {
	return &FS{
		dir: dir,
	}
}

func (fs *FS) SaveGame(ctx context.Context, game *words.Game) (err error) {
	file := fs.file(game.ID)
	err = os.MkdirAll(fs.dir, os.ModePerm)
	if err != nil {
		return err
	}

	// open file for writing
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, f.Close())
	}()

	compressor := gzip.NewWriter(f)
	defer func() {
		err = errors.Join(err, compressor.Close())
	}()

	err = json.NewEncoder(compressor).Encode(game)
	if err != nil {
		return err
	}

	return
}

func (fs *FS) file(id string) string {
	return filepath.Join(fs.dir, id+".json.gz")
}

func (fs *FS) GetGameByID(ctx context.Context, id string) (game *words.Game, err error) {
	file := fs.file(id)

	// stream read file
	f, err := os.Open(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() {
		err = errors.Join(err, f.Close())
	}()

	decompressor, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("error creating decompressor: %w", err)
	}
	defer func() {
		err = errors.Join(err, decompressor.Close())
	}()

	game = &words.Game{}
	err = json.NewDecoder(decompressor).Decode(game)
	if err != nil {
		return nil, err
	}

	return
}
