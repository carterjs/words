package game

import (
	"context"
	"errors"
	"log/slog"
)

type (
	Service struct {
		logger *slog.Logger
		store  Store
	}

	Store interface {
		SaveGame(ctx context.Context, game Game) error
		GetGameByID(ctx context.Context, id string) (*Game, error)
		SaveTurn(ctx context.Context, turn Turn) error
		GetTurnsByGameID(ctx context.Context, gameID string) ([]Turn, error)
		SavePlayer(ctx context.Context, player Player) error
		GetPlayersByGameID(ctx context.Context, gameID string) ([]Player, error)
		GetPlayerByID(ctx context.Context, playerID string) (*Player, error)
	}
)

func NewService(logger *slog.Logger, store Store) *Service {
	return &Service{
		logger: logger,
		store:  store,
	}
}

func (service *Service) CreateGame(ctx context.Context, name string, passphrase string) (*Game, error) {
	game := New(name, passphrase)

	err := service.store.SaveGame(ctx, *game)
	if err != nil {
		return nil, err
	}

	return game, nil
}

func (service *Service) GetGameByID(ctx context.Context, id string) (*Game, error) {
	return service.store.GetGameByID(ctx, id)
}

func (service *Service) AddPlayerToGame(ctx context.Context, gameID string, name string, passphrase string) (*Player, error) {
	game, err := service.GetGameByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if !game.PassphraseMatches(passphrase) {
		return nil, ErrIncorrectPassphrase
	}

	player := NewPlayer(gameID, name)
	err = service.store.SavePlayer(ctx, *player)
	if err != nil {
		return nil, err
	}

	return player, nil
}

func (service *Service) GetPlayersByGameID(ctx context.Context, gameID string) ([]Player, error) {
	return service.store.GetPlayersByGameID(ctx, gameID)
}

func (service *Service) GetPlayerByID(ctx context.Context, playerID string) (*Player, error) {
	return service.store.GetPlayerByID(ctx, playerID)
}

func (service *Service) GiveLettersToPlayer(ctx context.Context, playerID string, letters []rune) error {
	player, err := service.store.GetPlayerByID(ctx, playerID)
	if err != nil {
		return err
	}

	player.Letters = append(player.Letters, letters...)

	err = service.store.SavePlayer(ctx, *player)
	if err != nil {
		return err
	}

	return nil
}

func (service *Service) GetBoard(ctx context.Context, gameID string) (Board, error) {
	// get all turns, construct a board
	// TODO: cache or improve querying somehow
	turns, err := service.store.GetTurnsByGameID(context.Background(), gameID)
	if err != nil {
		return Board{}, err
	}

	var words []Word
	for _, turn := range turns {
		words = append(words, turn.Word)
	}

	board, err := NewBoard(words)
	if err != nil {
		return Board{}, err
	}

	return board, nil
}

type TurnRequest struct {
	GameID   string
	PlayerID string
	Word     Word
}

func (service *Service) RecordTurn(ctx context.Context, request TurnRequest) (*Turn, error) {
	turn := NewTurn(request.GameID, request.PlayerID, request.Word)

	player, err := service.GetPlayerByID(ctx, request.PlayerID)
	if err != nil {
		return nil, err
	}

	if player.HasLettersForWord(request.Word.Value) {
		return nil, errors.New("player does not have the letters to play the word")
	}

	// validate the word

	// remove those letters from the player's letters
	// verify that the word fits among the other words
	// determine points

	err = service.store.SaveTurn(ctx, *turn)
	if err != nil {
		return nil, err
	}

	return turn, nil
}
