package game

import (
	"context"
	"errors"
	"log/slog"
)

type (
	Service struct {
		logger *slog.Logger
		store  ServiceStore
	}

	ServiceStore interface {
		GameStore
		TurnStore
		VoteStore
		PlayerStore
		UsageStore
	}

	GameStore interface {
		CreateGame(ctx context.Context, game Game) error
		UpdateGame(ctx context.Context, game Game) error
		GetGameByID(ctx context.Context, id string) (*Game, error)
	}

	TurnStore interface {
		CreateTurn(ctx context.Context, turn Turn) error
		UpdateTurn(ctx context.Context, turn Turn) error
		GetTurnByID(ctx context.Context, id string) (*Turn, error)
		GetTurnsByGameID(ctx context.Context, gameID string) ([]Turn, error)
		GetTurnsByRound(ctx context.Context, gameID string, round int) ([]Turn, error)
	}

	VoteStore interface {
		CreateTurnVote(ctx context.Context, turnVote TurnVote) error
		GetTurnVotes(ctx context.Context, turnID string) ([]TurnVote, error)
	}

	PlayerStore interface {
		CreatePlayer(ctx context.Context, player Player) error
		UpdatePlayer(ctx context.Context, player Player) error
		GetPlayersByGameID(ctx context.Context, gameID string) ([]Player, error)
		GetPlayerByID(ctx context.Context, playerID string) (*Player, error)
	}

	UsageStore interface {
		GetWordStats(ctx context.Context, word string) (WordStats, error)
		SaveWordUsage(ctx context.Context, word string, approvals, rejections int) error
	}
)

func NewService(logger *slog.Logger, store ServiceStore) *Service {
	return &Service{
		logger: logger,
		store:  store,
	}
}

func (service *Service) CreateGame(ctx context.Context, name string, passphrase string) (*Game, error) {
	game := New(name, passphrase)

	err := service.store.CreateGame(ctx, *game)
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
		if errors.Is(err, ErrGameNotFound) {
			return nil, ErrGameNotFound
		}

		return nil, err
	}

	if !game.PassphraseMatches(passphrase) {
		return nil, ErrIncorrectPassphrase
	}

	player := NewPlayer(gameID, name)
	err = service.store.CreatePlayer(ctx, *player)
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
	player, err := service.GetPlayerByID(ctx, playerID)
	if err != nil {
		return err
	}

	player.Letters = append(player.Letters, letters...)

	err = service.store.UpdatePlayer(ctx, *player)
	if err != nil {
		return err
	}

	return nil
}

func (service *Service) GetBoard(ctx context.Context, gameID string) (*Board, error) {
	// get all turns, construct a board
	// TODO: cache or improve querying somehow
	turns, err := service.store.GetTurnsByGameID(context.Background(), gameID)
	if err != nil {
		return nil, err
	}

	var words []Word
	for _, turn := range turns {
		words = append(words, turn.Word)
	}

	board, err := NewBoard(words)
	if err != nil {
		return nil, err
	}

	return board, nil
}

func (service *Service) SubmitTurn(ctx context.Context, gameID string, playerID string, word Word) (*Turn, error) {
	game, err := service.GetGameByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	// get all players
	players, err := service.GetPlayersByGameID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	player, inGame := getPlayer(players, playerID)
	if !inGame {
		return nil, ErrPlayerNotFound
	}

	turn := NewTurn(gameID, game.Round, playerID, word)

	// single player games get auto approval
	if len(players) == 1 {
		turn.Status = TurnStatusPlayed
	}

	// TODO: auto approve if the word is recognized
	wordStats, err := service.store.GetWordStats(ctx, turn.Word.String())
	if err != nil {
		return nil, err
	}

	// auto approve with high reputation
	if wordStats.Reputation() > 0.8 {
		turn.Status = TurnStatusPlayed
	}

	board, err := service.GetBoard(ctx, gameID)
	if err != nil {
		return nil, err
	}

	wordPlacement, err := board.TryWordPlacement(word)
	if err != nil {
		return nil, err
	}

	if player.HasLetters(wordPlacement.LettersUsed) {
		return nil, errors.New("player does not have the letters to play the word")
	}

	err = service.store.CreateTurn(ctx, *turn)
	if err != nil {
		return nil, err
	}

	if turn.Status == TurnStatusPlayed {
		// if the status was auto approved somehow, see if we should advance the round
		service.updateGameState(ctx, gameID, players)
	}

	return turn, nil
}

func getPlayer(players []Player, playerID string) (*Player, bool) {
	for _, player := range players {
		if player.ID == playerID {
			return &player, true
		}
	}

	return nil, false
}

func (service *Service) VoteOnTurn(ctx context.Context, gameID string, turnID string, playerID string, approved bool) error {
	// check that it's a real turn
	turn, err := service.store.GetTurnByID(ctx, turnID)
	if err != nil {
		return err
	}

	if playerID == turn.PlayerID {
		return ErrSelfVote
	}

	// check that the player is in the associated game
	if ok, err := service.playerIsInGame(ctx, gameID, playerID); err != nil {
		return err
	} else if !ok {
		return ErrPlayerNotFound
	}

	var voteValue TurnVoteValue
	if approved {
		voteValue = TurnVoteValueApprove
	} else {
		voteValue = TurnVoteValueReject
	}

	err = service.store.CreateTurnVote(ctx, TurnVote{
		TurnID:   turnID,
		PlayerID: playerID,
		Value:    voteValue,
	})
	if err != nil {
		return err
	}

	players, err := service.GetPlayersByGameID(ctx, gameID)
	if err != nil {
		return err
	}

	// update the turn status if necessary
	err = service.updateTurnStatus(ctx, *turn, players)
	if err != nil {
		return err
	}

	err = service.updateGameState(ctx, gameID, players)
	if err != nil {
		return err
	}

	return nil
}

func (service *Service) playerIsInGame(ctx context.Context, gameID string, playerID string) (bool, error) {
	players, err := service.GetPlayersByGameID(ctx, gameID)
	if err != nil {
		return false, err
	}

	for _, player := range players {
		if player.ID == playerID {
			return true, nil
		}
	}

	return false, nil
}

func (service *Service) updateTurnStatus(ctx context.Context, turn Turn, players []Player) error {
	var activePlayerCount int
	for _, player := range players {
		if player.Status == PlayerStatusActive {
			activePlayerCount++
		}
	}

	votes, err := service.store.GetTurnVotes(ctx, turn.ID)
	if err != nil {
		return err
	}

	var votesByValue = make(map[TurnVoteValue]int)
	for _, vote := range votes {
		votesByValue[vote.Value]++
	}

	if votesByValue[TurnVoteValueApprove] <= (activePlayerCount-1)/2 {
		// not approved yet
		return nil
	}

	// approved!
	turn.Status = TurnStatusPlayed
	err = service.store.UpdateTurn(ctx, turn)
	if err != nil {
		return err
	}
	err = service.store.SaveWordUsage(ctx, turn.Word.String(), votesByValue[TurnVoteValueApprove], votesByValue[TurnVoteValueReject])
	if err != nil {
		return err
	}

	return nil
}

func (service *Service) updateGameState(ctx context.Context, gameID string, players []Player) error {
	game, err := service.GetGameByID(ctx, gameID)
	if err != nil {
		return err
	}

	turns, err := service.store.GetTurnsByRound(ctx, gameID, game.Round)
	if err != nil {
		return err
	}

	var playersWithTurns = make(map[string]bool)
	for _, turn := range turns {
		turnPlayed := turn.Status == TurnStatusPlayed
		playersWithTurns[turn.PlayerID] = turnPlayed
	}

	for _, player := range players {
		if player.Status == PlayerStatusInactive {
			continue
		}

		if !playersWithTurns[player.ID] {
			return nil
		}
	}

	// increment round
	game.Round++
	err = service.store.UpdateGame(ctx, *game)
	if err != nil {
		return err
	}

	return nil
}
