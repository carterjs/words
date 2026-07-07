package words

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
)

// Store persists games between requests. Implementations translate their
// own failures into this package's errors, notably ErrGameNotFound.
type Store interface {
	SaveGame(ctx context.Context, game *Game) error
	GameByID(ctx context.Context, gameID string) (*Game, error)
}

// Broker fans events out to game subscribers.
type Broker interface {
	Publish(ctx context.Context, channel string, event Event)
	Subscribe(ctx context.Context, channels ...string) Subscription
}

// Subscription is one subscriber's stream of game events.
type Subscription interface {
	Next(ctx context.Context) (Event, error)
	Close()
}

// Service coordinates game rules, persistence, and event delivery. Every
// game mutation goes through it, so concurrent requests against the same
// game are serialized.
type Service struct {
	store  Store
	broker Broker
	logger *slog.Logger

	mutex     sync.Mutex
	gameLocks map[string]*sync.Mutex
}

// NewService returns a service backed by the given store and broker.
func NewService(store Store, broker Broker, logger *slog.Logger) *Service {
	return &Service{
		store:     store,
		broker:    broker,
		logger:    logger,
		gameLocks: make(map[string]*sync.Mutex),
	}
}

// CreateGame creates and persists a new game from the given preset.
func (service *Service) CreateGame(ctx context.Context, presetID string, overrides ConfigOverrides) (*Game, error) {
	preset, exists := PresetByID(presetID)
	if !exists {
		return nil, ErrPresetNotFound
	}

	game := NewGame(configWithOverrides(preset.Config, overrides))

	if err := service.store.SaveGame(ctx, game); err != nil {
		return nil, fmt.Errorf("saving new game: %w", err)
	}

	return game, nil
}

// GameByID returns the game with the given ID.
func (service *Service) GameByID(ctx context.Context, gameID string) (*Game, error) {
	game, err := service.store.GameByID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("loading game: %w", err)
	}

	return game, nil
}

// JoinGame adds a player to the game and announces them to subscribers.
func (service *Service) JoinGame(ctx context.Context, gameID, playerName string) (*Game, Player, error) {
	defer service.lockGame(gameID)()

	game, err := service.GameByID(ctx, gameID)
	if err != nil {
		return nil, Player{}, fmt.Errorf("joining game: %w", err)
	}

	player, err := game.AddPlayer(playerName)
	if err != nil {
		return nil, Player{}, fmt.Errorf("adding player: %w", err)
	}

	if err := service.store.SaveGame(ctx, game); err != nil {
		return nil, Player{}, fmt.Errorf("saving game: %w", err)
	}

	service.publish(ctx, gameChannel(gameID), EventTypePlayerJoined, PlayerJoinedPayload{
		PlayerID:   player.ID(),
		PlayerName: player.Name(),
	})

	return game, player, nil
}

// StartGame starts the game and deals every player their opening rack.
func (service *Service) StartGame(ctx context.Context, gameID string) (*Game, error) {
	defer service.lockGame(gameID)()

	game, err := service.GameByID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("loading game for turn: %w", err)
	}

	if err := game.Start(); err != nil {
		return nil, fmt.Errorf("starting game: %w", err)
	}

	if err := service.store.SaveGame(ctx, game); err != nil {
		return nil, fmt.Errorf("saving game: %w", err)
	}

	service.publish(ctx, gameChannel(gameID), EventTypeGameStarted, GameStartedPayload{})
	for _, player := range game.Players() {
		service.publish(ctx, playerChannel(gameID, player.ID()), EventTypeGameStarted, GameStartedPayload{
			Letters: letterStrings(player.Letters()),
		})
	}

	return game, nil
}

// PlayWord plays a word for the given player and broadcasts the result.
func (service *Service) PlayWord(ctx context.Context, gameID, playerID string, word Word) (*Game, PlacementResult, error) {
	defer service.lockGame(gameID)()

	game, err := service.GameByID(ctx, gameID)
	if err != nil {
		return nil, PlacementResult{}, fmt.Errorf("loading game for play: %w", err)
	}

	result, err := game.PlayWord(playerID, word)
	if err != nil {
		return nil, PlacementResult{}, fmt.Errorf("playing word: %w", err)
	}

	if err := service.store.SaveGame(ctx, game); err != nil {
		return nil, PlacementResult{}, fmt.Errorf("saving game: %w", err)
	}

	start := result.DirectWord.Start()
	service.publish(ctx, gameChannel(gameID), EventTypeWordPlayed, WordPlayedPayload{
		PlayerID:     playerID,
		X:            start.Column(),
		Y:            start.Row(),
		Direction:    result.DirectWord.Direction(),
		Word:         string(result.DirectWord.Letters()),
		Points:       result.Points,
		NextPlayerID: game.CurrentPlayerID(),
		Round:        game.Round(),
	})
	service.publishRack(ctx, game, playerID)
	service.publishGameEndedIfFinished(ctx, game)

	return game, result, nil
}

// PassTurn forfeits the given player's turn and broadcasts it.
func (service *Service) PassTurn(ctx context.Context, gameID, playerID string) (*Game, error) {
	defer service.lockGame(gameID)()

	game, err := service.GameByID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("loading game for turn: %w", err)
	}

	if err := game.PassTurn(playerID); err != nil {
		return nil, fmt.Errorf("passing turn: %w", err)
	}

	if err := service.store.SaveGame(ctx, game); err != nil {
		return nil, fmt.Errorf("saving game: %w", err)
	}

	service.publish(ctx, gameChannel(gameID), EventTypeTurnPassed, TurnPassedPayload{
		PlayerID:     playerID,
		NextPlayerID: game.CurrentPlayerID(),
		Round:        game.Round(),
	})
	service.publishGameEndedIfFinished(ctx, game)

	return game, nil
}

// ExchangeLetters swaps letters for the given player and broadcasts the
// exchange without revealing the letters.
func (service *Service) ExchangeLetters(ctx context.Context, gameID, playerID string, letters []rune) (*Game, error) {
	defer service.lockGame(gameID)()

	game, err := service.GameByID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("loading game for turn: %w", err)
	}

	if err := game.ExchangeLetters(playerID, letters); err != nil {
		return nil, fmt.Errorf("exchanging letters: %w", err)
	}

	if err := service.store.SaveGame(ctx, game); err != nil {
		return nil, fmt.Errorf("saving game: %w", err)
	}

	service.publish(ctx, gameChannel(gameID), EventTypeLettersExchanged, LettersExchangedPayload{
		PlayerID:     playerID,
		Count:        len(letters),
		NextPlayerID: game.CurrentPlayerID(),
		Round:        game.Round(),
	})
	service.publishRack(ctx, game, playerID)
	service.publishGameEndedIfFinished(ctx, game)

	return game, nil
}

// ChallengeWord opens a consensus vote against the last played word.
func (service *Service) ChallengeWord(ctx context.Context, gameID, playerID string) (*Game, ChallengeOutcome, error) {
	defer service.lockGame(gameID)()

	game, err := service.GameByID(ctx, gameID)
	if err != nil {
		return nil, ChallengeOutcome{}, fmt.Errorf("loading game for challenge: %w", err)
	}

	outcome, err := game.Challenge(playerID)
	if err != nil {
		return nil, ChallengeOutcome{}, fmt.Errorf("challenging word: %w", err)
	}

	if err := service.store.SaveGame(ctx, game); err != nil {
		return nil, ChallengeOutcome{}, fmt.Errorf("saving game: %w", err)
	}

	service.publish(ctx, gameChannel(gameID), EventTypeChallengeStarted, ChallengeStartedPayload{
		ChallengerID:   outcome.ChallengerID,
		MoverID:        outcome.MoverID,
		VotesInvalid:   outcome.VotesInvalid,
		VotesValid:     outcome.VotesValid,
		VotesNeeded:    outcome.VotesNeeded,
		EligibleVoters: outcome.EligibleVoters,
	})
	service.publishChallengeResolution(ctx, game, outcome)

	return game, outcome, nil
}

// CastVote adds a player's vote to the open challenge and broadcasts the
// tally, resolving the challenge once the outcome is decided.
func (service *Service) CastVote(ctx context.Context, gameID, playerID string, vote Vote) (*Game, ChallengeOutcome, error) {
	defer service.lockGame(gameID)()

	game, err := service.GameByID(ctx, gameID)
	if err != nil {
		return nil, ChallengeOutcome{}, fmt.Errorf("loading game for challenge: %w", err)
	}

	outcome, err := game.CastVote(playerID, vote)
	if err != nil {
		return nil, ChallengeOutcome{}, fmt.Errorf("casting vote: %w", err)
	}

	if err := service.store.SaveGame(ctx, game); err != nil {
		return nil, ChallengeOutcome{}, fmt.Errorf("saving game: %w", err)
	}

	service.publish(ctx, gameChannel(gameID), EventTypeChallengeVoteCast, ChallengeVoteCastPayload{
		PlayerID:     playerID,
		VotesInvalid: outcome.VotesInvalid,
		VotesValid:   outcome.VotesValid,
		VotesNeeded:  outcome.VotesNeeded,
	})
	service.publishChallengeResolution(ctx, game, outcome)

	return game, outcome, nil
}

// Subscribe returns the stream of events for a game, including the private
// events of the given player when playerID is not empty.
func (service *Service) Subscribe(ctx context.Context, gameID, playerID string) Subscription {
	return service.broker.Subscribe(ctx, gameChannel(gameID), playerChannel(gameID, playerID))
}

func (service *Service) publishChallengeResolution(ctx context.Context, game *Game, outcome ChallengeOutcome) {
	if !outcome.Resolved {
		return
	}

	payload := ChallengeResolvedPayload{
		Upheld:       outcome.Upheld,
		ChallengerID: outcome.ChallengerID,
		MoverID:      outcome.MoverID,
		VotesInvalid: outcome.VotesInvalid,
		VotesValid:   outcome.VotesValid,
	}
	if outcome.RescindedWord != nil {
		payload.RescindedWord = string(outcome.RescindedWord.Letters())
	}

	service.publish(ctx, gameChannel(game.ID()), EventTypeChallengeResolved, payload)

	if outcome.Upheld {
		service.publishRack(ctx, game, outcome.MoverID)
	}
}

func (service *Service) publishGameEndedIfFinished(ctx context.Context, game *Game) {
	if !game.Finished() {
		return
	}

	scores := make(map[string]int)
	for _, player := range game.Players() {
		scores[player.ID()] = player.Score()
	}

	service.publish(ctx, gameChannel(game.ID()), EventTypeGameEnded, GameEndedPayload{
		WinnerIDs: game.WinnerIDs(),
		Scores:    scores,
	})
}

func (service *Service) publishRack(ctx context.Context, game *Game, playerID string) {
	player, exists := game.PlayerByID(playerID)
	if !exists {
		return
	}

	service.publish(ctx, playerChannel(game.ID(), playerID), EventTypeRackUpdated, RackUpdatedPayload{
		Letters: letterStrings(player.Letters()),
	})
}

func (service *Service) publish(ctx context.Context, channel string, eventType EventType, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		service.logger.Error("marshaling event payload", "type", eventType, "error", err)
		return
	}

	service.broker.Publish(ctx, channel, Event{Type: eventType, Payload: data})
}

// lockGame serializes mutations per game and returns the unlock function.
func (service *Service) lockGame(gameID string) func() {
	service.mutex.Lock()
	lock, exists := service.gameLocks[gameID]
	if !exists {
		lock = &sync.Mutex{}
		service.gameLocks[gameID] = lock
	}
	service.mutex.Unlock()

	lock.Lock()
	return lock.Unlock
}

func gameChannel(gameID string) string {
	return "game:" + gameID
}

func playerChannel(gameID, playerID string) string {
	return "game:" + gameID + ":player:" + playerID
}

func letterStrings(letters []rune) []string {
	strings := make([]string, len(letters))
	for index, letter := range letters {
		strings[index] = string(letter)
	}

	return strings
}
