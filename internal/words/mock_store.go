package words

import "context"

// MockStore is a hand-written functional mock of Store for tests. A nil
// function field panics to surface unexpected calls.
type MockStore struct {
	SaveGameFunc func(ctx context.Context, game *Game) error
	GameByIDFunc func(ctx context.Context, gameID string) (*Game, error)
}

// SaveGame calls SaveGameFunc.
func (mock *MockStore) SaveGame(ctx context.Context, game *Game) error {
	return mock.SaveGameFunc(ctx, game)
}

// GameByID calls GameByIDFunc.
func (mock *MockStore) GameByID(ctx context.Context, gameID string) (*Game, error) {
	return mock.GameByIDFunc(ctx, gameID)
}
