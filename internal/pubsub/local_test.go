package pubsub_test

import (
	"testing"

	"github.com/carterjs/words/internal/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocal_Publish(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		subscribeKey string
		publishKeys  []string
		want         string
	}{
		{name: "delivers to a matching subscription", subscribeKey: "a", publishKeys: []string{"a"}, want: "value:a"},
		{name: "skips values published on other keys", subscribeKey: "a", publishKeys: []string{"b", "a"}, want: "value:a"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			local := pubsub.NewLocal[string, string]()
			subscription := local.Subscribe(test.subscribeKey)
			defer subscription.Close()

			for _, key := range test.publishKeys {
				local.Publish(t.Context(), key, "value:"+key)
			}

			value, err := subscription.Next(t.Context())
			require.NoError(t, err)
			assert.Equal(t, test.want, value)
		})
	}
}

func TestSubscription_Close(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "next reports a closed subscription"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			local := pubsub.NewLocal[string, string]()
			subscription := local.Subscribe("a")

			subscription.Close()
			subscription.Close() // closing twice is safe

			_, err := subscription.Next(t.Context())
			assert.ErrorIs(t, err, pubsub.ErrSubscriptionClosed)

			// publishing after close does not panic or block
			local.Publish(t.Context(), "a", "late")
		})
	}
}
