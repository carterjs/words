package server

import (
	"encoding/json"
	"fmt"
	"github.com/carterjs/words/internal/words"
	"github.com/gorilla/websocket"
	"log/slog"
	"net/http"
)

type request interface {
	Execute(*Server, *websocket.Conn) error
}

type message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type playerStatusChange struct {
	PlayerID string `json:"playerId"`
	Online   bool   `json:"online"`
}

type errorMessage struct {
	Message string `json:"message"`
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
			defer func() {
				err := conn.Close()
				if err != nil {
					slog.Error("error closing connection", "error", err)
				}
				sess := server.getSession(conn)
				server.deleteSession(conn)

				if sess.gameID == "" {
					return
				}

				err = server.broadcastResponse(sess.gameID, func(session) (string, any) {
					return "player_offline", playerStatusChange{
						PlayerID: server.getSession(conn).playerID,
						Online:   false,
					}
				})
				if err != nil {
					slog.Error("error broadcasting player offline message", "error", err)
				}
			}()

			for {
				var msg message
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

func (server *Server) handleCommand(conn *websocket.Conn, msg message) error {
	switch msg.Type {
	case "create_game":
		return execute[createGameRequest](server, conn, msg.Payload)
	case "join_game":
		return execute[joinGameRequest](server, conn, msg.Payload)
	case "rejoin_game":
		return execute[rejoinGameRequest](server, conn, msg.Payload)
	case "start_game":
		return execute[startGameRequest](server, conn, msg.Payload)
	case "play_word":
		return execute[playWordRequest](server, conn, msg.Payload)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

func execute[T request](server *Server, conn *websocket.Conn, payload json.RawMessage) error {
	var req T
	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("error unmarshalling request: %w", err)
	}

	err := req.Execute(server, conn)
	if err != nil {
		return server.sendResponse(conn, "error", errorMessage{
			Message: err.Error(),
		})
	}

	return nil
}

func (server *Server) sendResponse(conn *websocket.Conn, messageType string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return conn.WriteJSON(message{
		Type:    messageType,
		Payload: b,
	})
}

func (server *Server) broadcastResponse(gameID string, get func(session) (string, any)) error {
	server.mu.Lock()
	defer server.mu.Unlock()

	for conn, sess := range server.connections {
		if sess.gameID == gameID {
			t, payload := get(sess)
			if err := server.sendResponse(conn, t, payload); err != nil {
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
			letter, isSet := game.Board.GetLetter(words.NewPoint(x, y))
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
