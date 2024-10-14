package main

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/carterjs/words/internal/words"
	"github.com/gorilla/websocket"
)

type (
	Server struct {
		store GameStore

		mu          sync.Mutex
		connections map[*websocket.Conn]session
		publicDir   string
	}

	GameStore interface {
		SaveGame(ctx context.Context, game *words.Game) error
		GetGameByID(ctx context.Context, id string) (*words.Game, error)
	}

	session struct {
		gameID   string
		playerID string
	}
)

const (
	minimumBoardSize = 8
)

func NewServer(store GameStore, publicDir string) *Server {
	return &Server{
		store:       store,
		connections: make(map[*websocket.Conn]session),
		publicDir:   publicDir,
	}
}

func (server *Server) Handler() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /health", server.handleGetHealth())
	mux.Handle("GET /ws", server.handleWS())
	mux.Handle("GET /", server.handlePublic())

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

func (server *Server) handlePublic() http.Handler {
	fs := http.FileServer(http.Dir(server.publicDir))

	return fs
	// log.Println("serving files from", server.publicDir)

	// return func(writer http.ResponseWriter, r *http.Request) {
	// 	log.Println("serving", r.URL.Path)
	// 	// if it has a file extension, serve file
	// 	if strings.Contains(r.URL.Path, ".") || strings.Count(r.URL.Path, "/") > 1 {
	// 		fs.ServeHTTP(writer, r)
	// 		return
	// 	}

	// 	// otherwise, serve index.html
	// 	http.ServeFile(writer, r, "public/index.html")
	// }
}
