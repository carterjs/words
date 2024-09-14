package server

import (
	"context"
	"encoding/json"
	"github.com/carterjs/words/internal/store"
	"github.com/carterjs/words/internal/words"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"
	"os"
	"sync"
)

type (
	Server struct {
		store GameStore

		mu          sync.Mutex
		connections map[*websocket.Conn]session
	}

	GameStore interface {
		SaveGame(ctx context.Context, game *words.Game) error
		GetGameByID(ctx context.Context, id string) (*words.Game, error)
	}

	session struct {
		ws       *websocket.Conn
		gameID   string
		playerID string
	}
)

func Command() *cobra.Command {
	viper.SetDefault("port", "8080")
	viper.SetDefault("directory", "/tmp/word-game")
	viper.MustBindEnv("port", "PORT")
	viper.MustBindEnv("directory", "DIR")

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	cmd := &cobra.Command{
		Use: "server",
		RunE: func(cmd *cobra.Command, args []string) error {
			port := viper.GetString("port")
			directory := viper.GetString("directory")

			server := NewServer(store.NewFS(directory))
			logger.Info("starting server", "port", port)
			return http.ListenAndServe(":"+port, server.Handler())
		},
	}

	return cmd
}

func NewServer(store GameStore) *Server {
	return &Server{
		store:       store,
		connections: make(map[*websocket.Conn]session),
	}
}

func (server *Server) Handler() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /health", server.handleGetHealth())
	mux.Handle("GET /ws", server.handleWS())

	return mux
}

func (server *Server) handleGetHealth() http.HandlerFunc {
	type response struct {
		Status string `json:"status"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response{Status: "ok"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
