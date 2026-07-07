package pubsub_test

import (
	"encoding/json"
	"testing"

	"github.com/carterjs/words/internal/pubsub"
	"github.com/carterjs/words/internal/words"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGameBroker_Publish(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		subscribeTo    []string
		publishChannel string
		event          words.Event
	}{
		{
			name:           "delivers game events across channels",
			subscribeTo:    []string{"game:1", "game:1:player:2"},
			publishChannel: "game:1:player:2",
			event:          words.Event{Type: words.EventTypeRackUpdated, Payload: json.RawMessage(`{"letters":[]}`)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			broker := pubsub.NewGameBroker()
			subscription := broker.Subscribe(t.Context(), test.subscribeTo...)
			defer subscription.Close()

			broker.Publish(t.Context(), test.publishChannel, test.event)

			received, err := subscription.Next(t.Context())
			require.NoError(t, err)
			assert.Equal(t, test.event, received)
		})
	}
}
