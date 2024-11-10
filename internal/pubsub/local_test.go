package pubsub_test

import (
	"context"
	"github.com/carterjs/words/internal/pubsub"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocal(t *testing.T) {
	local := pubsub.NewLocal[string, string]()

	key1, unsubscribe := local.Subscribe(context.Background(), "key1")
	defer unsubscribe()

	key2, unsubscribe := local.Subscribe(context.Background(), "key2")
	defer unsubscribe()

	go local.Publish(context.Background(), "key1", "value1")
	go local.Publish(context.Background(), "key2", "value2")
	assert.Equal(t, "value1", <-key1)
	assert.Equal(t, "value2", <-key2)
}
