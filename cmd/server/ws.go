package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/carterjs/words/internal/words"
	"github.com/gorilla/websocket"
	"log/slog"
	"net/http"
)

type message interface {
	Type() string
}

type rawMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type createGameRequest struct {
	PlayerName string `json:"playerName"`
}

func (createGameRequest) Type() string {
	return "create_game"
}

type playerInfoResponse struct {
	PlayerID string `json:"playerId"`
	GameID   string `json:"gameId"`
}

func (playerInfoResponse) Type() string {
	return "player_info"
}

type joinGameRequest struct {
	PlayerName string `json:"playerName"`
	GameID     string `json:"gameId"`
}

func (joinGameRequest) Type() string {
	return "join_game"
}

type startGameRequest struct{}

func (startGameRequest) Type() string {
	return "start_game"
}

type (
	startGameResponse struct {
		Players []playerInfo           `json:"players"`
		Turn    int                    `json:"turn"`
		Grid    map[words.Point]string `json:"grid"`
		Rack    []string               `json:"rack"`
	}

	playerInfo struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
)

func (startGameResponse) Type() string {
	return "start_game"
}

type playWordRequest struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Direction string `json:"direction"`
	Word      string `json:"word"`
}

func (playWordRequest) Type() string {
	return "play_word"
}

type playWordResponse struct {
	Word   string                 `json:"word"`
	Grid   map[words.Point]string `json:"grid"`
	Points int                    `json:"points"`
}

func (playWordResponse) Type() string {
	return "play_word"
}

func (server *Server) handleWS() http.HandlerFunc {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		go func() {
			defer conn.Close()
			defer server.deleteSession(conn)

			for {
				var msg rawMessage
				err := conn.ReadJSON(&msg)
				if err != nil {
					slog.Error("error reading json message", "error", err)
					return
				}

				err = server.handleCommand(conn, msg)
				if err != nil {
					slog.Error("error handling message", "type", msg.Type, "error", err)
					return
				}
			}
		}()

		server.saveSession(conn, session{})
	}
}

func (server *Server) saveSession(conn *websocket.Conn, value session) {
	server.mu.Lock()
	defer server.mu.Unlock()

	server.connections[conn] = value
}

func (server *Server) getSession(conn *websocket.Conn) session {
	server.mu.Lock()
	defer server.mu.Unlock()

	return server.connections[conn]
}

func (server *Server) deleteSession(conn *websocket.Conn) {
	server.mu.Lock()
	defer server.mu.Unlock()

	delete(server.connections, conn)
}

func (server *Server) handleCommand(conn *websocket.Conn, cmd rawMessage) error {
	switch cmd.Type {
	case createGameRequest{}.Type():
		var createGame createGameRequest
		err := json.Unmarshal(cmd.Payload, &createGame)
		if err != nil {
			return err
		}

		return server.createGame(conn, createGame.PlayerName)
	case joinGameRequest{}.Type():
		var joinGame joinGameRequest
		err := json.Unmarshal(cmd.Payload, &joinGame)
		if err != nil {
			return err
		}

		return server.joinGame(conn, joinGame.GameID, joinGame.PlayerName)
	case startGameRequest{}.Type():
		return server.handleStartGame(conn)
	case playWordRequest{}.Type():
		var playWord playWordRequest
		err := json.Unmarshal(cmd.Payload, &playWord)
		if err != nil {
			return err
		}

		return server.handlePlayWord(conn, playWord)
	}

	return fmt.Errorf("unknown message type: %s", cmd.Type)
}

func (server *Server) sendResponse(conn *websocket.Conn, response message) error {
	b, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return conn.WriteJSON(rawMessage{
		Type:    response.Type(),
		Payload: b,
	})
}

func (server *Server) broadcastResponse(gameID string, response message) error {
	server.mu.Lock()
	defer server.mu.Unlock()

	for conn, sess := range server.connections {
		if sess.gameID == gameID {
			if err := server.sendResponse(conn, response); err != nil {
				return err
			}
		}
	}

	return nil
}

func (server *Server) createGame(conn *websocket.Conn, player string) error {
	game := words.NewGame(words.StandardConfig, player)
	err := server.store.SaveGame(context.Background(), game)
	if err != nil {
		return err
	}

	playerID := game.Players[0].ID

	server.saveSession(conn, session{gameID: game.ID, playerID: playerID})

	return server.sendResponse(conn, playerInfoResponse{
		PlayerID: game.Players[0].ID,
		GameID:   game.ID,
	})
}

func (server *Server) joinGame(conn *websocket.Conn, gameID string, playerID string) error {
	game, err := server.store.GetGameByID(context.Background(), gameID)
	if err != nil {
		return fmt.Errorf("error getting game: %w", err)
	}
	if game == nil {
		return fmt.Errorf("game not found: %s", gameID)
	}

	player, err := game.AddPlayer(playerID)
	if err != nil {
		return fmt.Errorf("error adding player: %w", err)
	}
	server.saveSession(conn, session{gameID: gameID, playerID: player.ID})

	err = server.store.SaveGame(context.Background(), game)
	if err != nil {
		return fmt.Errorf("error saving game: %w", err)
	}

	return server.sendResponse(conn, playerInfoResponse{PlayerID: player.ID, GameID: gameID})
}

func (server *Server) handleStartGame(conn *websocket.Conn) error {
	s := server.getSession(conn)
	if s.gameID == "" {
		return fmt.Errorf("no game to start")
	}

	game, err := server.store.GetGameByID(context.Background(), s.gameID)
	if err != nil {
		return fmt.Errorf("error getting game: %w", err)
	}
	if game == nil {
		return fmt.Errorf("game not found: %s", s.gameID)
	}

	if err := game.Start(); err != nil {
		return fmt.Errorf("error starting game: %w", err)
	}

	if err := server.store.SaveGame(context.Background(), game); err != nil {
		return fmt.Errorf("error saving game: %w", err)
	}

	return server.broadcastStartGame(game)
}

func (server *Server) broadcastStartGame(game *words.Game) error {
	server.mu.Lock()
	defer server.mu.Unlock()

	players := make([]playerInfo, len(game.Players))
	for i, p := range game.Players {
		players[i] = playerInfo{ID: p.ID, Name: p.Name}
	}

	for conn, sess := range server.connections {
		var letterRack []string
		for _, letter := range game.GetPlayerByID(sess.playerID).Letters {
			letterRack = append(letterRack, string(letter))
		}

		if sess.gameID == game.ID {
			if err := server.sendResponse(conn, startGameResponse{
				Players: players,
				Turn:    game.Turn,
				Grid:    getGrid(game, 0, 0, 14, 14),
				Rack:    letterRack,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func getGrid(game *words.Game, x1, y1, x2, y2 int) map[words.Point]string {
	grid := make(map[words.Point]string)

	for x := x1; x <= x2; x++ {
		for y := y1; y <= y2; y++ {
			letter, isSet := game.Board.GetLetter(x, y)
			if !isSet {
				modifier, isSet := game.Board.GetModifier(x, y)
				if isSet {
					grid[words.NewPoint(x, y)] = string(modifier)
				}

				continue
			}

			grid[words.NewPoint(x, y)] = string(letter)
		}
	}

	return grid
}

func (server *Server) handlePlayWord(conn *websocket.Conn, req playWordRequest) error {
	s := server.getSession(conn)
	if s.gameID == "" {
		return fmt.Errorf("no game to play word in")
	}

	game, err := server.store.GetGameByID(context.Background(), s.gameID)
	if err != nil {
		return fmt.Errorf("error getting game: %w", err)
	}
	if game == nil {
		return fmt.Errorf("game not found: %s", s.gameID)
	}

	direction := words.DirectionHorizontal
	if req.Direction == "vertical" {
		direction = words.DirectionVertical
	}

	word := words.NewWord(req.X, req.Y, direction, req.Word)
	result, err := game.PlayWord(s.playerID, word)
	if err != nil {
		return fmt.Errorf("error playing word: %w", err)
	}

	if err := server.store.SaveGame(context.Background(), game); err != nil {
		return fmt.Errorf("error saving game: %w", err)
	}

	lastX, lastY, _, _ := word.Index(len(word.Letters) - 1)
	return server.broadcastResponse(s.gameID, playWordResponse{
		Word:   req.Word,
		Grid:   getGrid(game, word.X, word.Y, lastX, lastY),
		Points: result.Points,
	})
}
