// Package pubsub provides in-process event fan-out: a generic broker plus an
// adapter satisfying the words service's Broker contract.
package pubsub

import (
	"context"
	"errors"
	"sync"
)

// ErrSubscriptionClosed reports that Next was called on a closed subscription.
var ErrSubscriptionClosed = errors.New("subscription closed")

// subscriptionBufferSize bounds how many undelivered events a subscriber can
// lag behind before further events are dropped for it.
const subscriptionBufferSize = 16

// Local is an in-process broker delivering values published on keys to the
// subscriptions listening on them.
type Local[K comparable, V any] struct {
	mutex         sync.RWMutex
	subscriptions map[*Subscription[V]][]K
}

// NewLocal returns an empty local broker.
func NewLocal[K comparable, V any]() *Local[K, V] {
	return &Local[K, V]{
		subscriptions: make(map[*Subscription[V]][]K),
	}
}

// Publish delivers the value to every subscription listening on the key.
// Delivery never blocks: a subscriber that has fallen more than the buffer
// size behind misses the value.
func (local *Local[K, V]) Publish(ctx context.Context, key K, value V) {
	local.mutex.RLock()
	defer local.mutex.RUnlock()

	for subscription, keys := range local.subscriptions {
		for _, subscribedKey := range keys {
			if subscribedKey != key {
				continue
			}

			select {
			case subscription.values <- value:
			case <-subscription.done:
			case <-ctx.Done():
				return
			default:
			}

			break
		}
	}
}

// Subscribe registers a new subscription for the given keys. The caller must
// Close it when done.
func (local *Local[K, V]) Subscribe(keys ...K) *Subscription[V] {
	subscription := &Subscription[V]{
		values: make(chan V, subscriptionBufferSize),
		done:   make(chan struct{}),
	}

	subscription.closeFunc = func() {
		local.remove(subscription)
		close(subscription.done)
	}

	local.mutex.Lock()
	defer local.mutex.Unlock()
	local.subscriptions[subscription] = keys

	return subscription
}

func (local *Local[K, V]) remove(subscription *Subscription[V]) {
	local.mutex.Lock()
	defer local.mutex.Unlock()

	delete(local.subscriptions, subscription)
}

// Subscription is one subscriber's stream of published values.
type Subscription[V any] struct {
	values    chan V
	done      chan struct{}
	closeOnce sync.Once
	closeFunc func()
}

// Next blocks until a value is delivered, the subscription is closed, or the
// context ends.
func (subscription *Subscription[V]) Next(ctx context.Context) (V, error) {
	select {
	case value := <-subscription.values:
		return value, nil
	case <-subscription.done:
		return *new(V), ErrSubscriptionClosed
	case <-ctx.Done():
		return *new(V), ctx.Err()
	}
}

// Close removes the subscription from its broker.
func (subscription *Subscription[V]) Close() {
	subscription.closeOnce.Do(subscription.closeFunc)
}
