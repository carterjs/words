package words_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/carterjs/words/internal/words"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_CreateGame(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		presetID string
		wantErr  error
	}{
		{name: "creates a game from a preset", presetID: "standard"},
		{name: "rejects an unknown preset", presetID: "nope", wantErr: words.ErrPresetNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var saved *words.Game
			service := newTestService(&words.MockStore{
				SaveGameFunc: func(ctx context.Context, game *words.Game) error {
					saved = game
					return nil
				},
			}, &words.MockBroker{})

			game, err := service.CreateGame(t.Context(), test.presetID, words.ConfigOverrides{RackSize: 3})

			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, game, saved)
			assert.Equal(t, 3, game.Config().RackSize)
		})
	}
}

func TestService_PlayWord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		outOfTurn      bool
		wantErr        error
		wantEventTypes []words.EventType
	}{
		{
			name:           "broadcasts the word and the player's new rack",
			wantEventTypes: []words.EventType{words.EventTypeWordPlayed, words.EventTypeRackUpdated},
		},
		{
			name:      "publishes nothing on a rejected play",
			outOfTurn: true,
			wantErr:   words.ErrNotYourTurn,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			game := newStartedGame(t, 2, testConfig(map[rune]int{'A': 20}, 3))
			service, published := newGameService(game)

			playerID := game.Players()[0].ID()
			if test.outOfTurn {
				playerID = game.Players()[1].ID()
			}

			_, result, err := service.PlayWord(t.Context(), game.ID(), playerID, horizontal(0, 0, "AA"))

			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
				assert.Empty(t, *published)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, 2, result.Points)
			assert.Equal(t, test.wantEventTypes, *published)
		})
	}
}

func newTestService(store *words.MockStore, broker *words.MockBroker) *words.Service {
	return words.NewService(store, broker, slog.New(slog.DiscardHandler))
}

// newGameService wires a service around one in-memory game, recording the
// types of every published event.
func newGameService(game *words.Game) (*words.Service, *[]words.EventType) {
	published := &[]words.EventType{}

	service := newTestService(
		&words.MockStore{
			GameByIDFunc: func(ctx context.Context, gameID string) (*words.Game, error) { return game, nil },
			SaveGameFunc: func(ctx context.Context, game *words.Game) error { return nil },
		},
		&words.MockBroker{
			PublishFunc: func(ctx context.Context, channel string, event words.Event) {
				*published = append(*published, event.Type)
			},
		},
	)

	return service, published
}
