package words

import "context"

// MockBroker is a hand-written functional mock of Broker for tests. A nil
// function field panics to surface unexpected calls.
type MockBroker struct {
	PublishFunc   func(ctx context.Context, channel string, event Event)
	SubscribeFunc func(ctx context.Context, channels ...string) Subscription
}

// Publish calls PublishFunc.
func (mock *MockBroker) Publish(ctx context.Context, channel string, event Event) {
	mock.PublishFunc(ctx, channel, event)
}

// Subscribe calls SubscribeFunc.
func (mock *MockBroker) Subscribe(ctx context.Context, channels ...string) Subscription {
	return mock.SubscribeFunc(ctx, channels...)
}
