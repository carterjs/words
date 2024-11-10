package pubsub

import (
	"context"
	"sync"
)

type Local[K comparable, V any] struct {
	mu sync.RWMutex
	ch map[chan V][]K
}

// NewLocal creates a new local pubsub broker
func NewLocal[K comparable, V any]() *Local[K, V] {
	return &Local[K, V]{
		ch: make(map[chan V][]K),
	}
}

// Publish sends the value to all subscribed channels
func (b *Local[K, V]) Publish(ctx context.Context, key K, value V) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch, keys := range b.ch {
		for _, k := range keys {
			if k == key {
				select {
				case ch <- value:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// Subscribe adds a new subscriber to the broker
func (b *Local[K, V]) Subscribe(ctx context.Context, keys ...K) (<-chan V, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan V)
	b.ch[ch] = keys

	return ch, func() {
		close(ch)
		b.unsubscribe(ch)
	}
}

func (b *Local[K, V]) unsubscribe(ch chan V) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.ch, ch)
}
