package pubsub

import (
	"context"

	"github.com/carterjs/words/internal/words"
)

// GameBroker adapts the generic local broker to the words service's Broker
// contract.
type GameBroker struct {
	local *Local[string, words.Event]
}

// NewGameBroker returns an in-process broker for game events.
func NewGameBroker() *GameBroker {
	return &GameBroker{
		local: NewLocal[string, words.Event](),
	}
}

// Publish delivers the event to every subscriber of the channel.
func (broker *GameBroker) Publish(ctx context.Context, channel string, event words.Event) {
	broker.local.Publish(ctx, channel, event)
}

// Subscribe returns a subscription covering all the given channels.
func (broker *GameBroker) Subscribe(ctx context.Context, channels ...string) words.Subscription {
	return broker.local.Subscribe(channels...)
}
