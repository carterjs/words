package main

// GameEvent is an interface that represents an event in the game
type GameEvent interface {
	Type() string
}

// GameStartedEvent is an event that represents the start of a game
type GameStartedEvent struct {
	Letters []string `json:"letters"`
}

func (GameStartedEvent) Type() string { return "GAME_STARTED" }

// MessageEvent is an event that represents a message sent by a player in the game
type MessageEvent struct {
	PlayerID string `json:"playerId"`
	Message  string `json:"message"`
}

func (MessageEvent) Type() string { return "MESSAGE" }
