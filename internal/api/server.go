// Package api serves the words service over HTTP: a JSON API under /api/v1
// plus the built frontend as static files. Errors are translated from the
// service's vocabulary to statuses here and nowhere else.
package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/carterjs/words/internal/errcode"
	"github.com/carterjs/words/internal/words"
)

const (
	defaultPublicDirectory = "public"
	defaultAllowedOrigin   = "http://localhost:5173"
)

// Config tunes the HTTP server.
type Config struct {
	PublicDirectory string
	AllowedOrigin   string
}

// Server exposes the words service over HTTP.
type Server struct {
	service *words.Service
	logger  *slog.Logger
	config  Config
}

// NewServer returns a server for the given service. Empty config fields fall
// back to defaults.
func NewServer(service *words.Service, logger *slog.Logger, config Config) *Server {
	if config.PublicDirectory == "" {
		config.PublicDirectory = defaultPublicDirectory
	}
	if config.AllowedOrigin == "" {
		config.AllowedOrigin = defaultAllowedOrigin
	}

	return &Server{
		service: service,
		logger:  logger,
		config:  config,
	}
}

// Handler returns the server's routes.
func (server *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// presets
	mux.Handle("GET /api/v1/presets", server.handleGetPresets())
	mux.Handle("GET /api/v1/presets/{id}", server.handleGetPresetByID())
	mux.Handle("GET /api/v1/presets/{id}/board", server.handleGetPresetBoard())

	// games
	mux.Handle("POST /api/v1/games", server.handleCreateGame())
	mux.Handle("GET /api/v1/games/{gameId}", server.handleGetGameByID())
	mux.Handle("PATCH /api/v1/games/{gameId}", server.handleUpdateGame())

	// board
	mux.Handle("GET /api/v1/games/{gameId}/board", server.handleGetGameBoard())
	mux.Handle("GET /api/v1/games/{gameId}/board/placements", server.handleGetGameBoardPlacements())
	mux.Handle("PATCH /api/v1/games/{gameId}/board", server.handleUpdateBoard())

	// events
	mux.Handle("GET /api/v1/games/{gameId}/events", server.handleStreamGameEvents())

	// frontend
	mux.Handle("/", http.FileServer(http.Dir(server.config.PublicDirectory)))

	return server.withCORS(mux)
}

func (server *Server) withCORS(handler http.Handler) http.Handler {
	methods := strings.Join([]string{http.MethodGet, http.MethodPost, http.MethodPatch}, ",")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", server.config.AllowedOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

// playerIDCookie carries the player's identity for a specific game.
const playerIDCookie = "playerId"

func playerIDFromRequest(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(playerIDCookie)
	if err != nil {
		return "", false
	}

	return cookie.Value, true
}

func parseRequestBody[T any](r *http.Request) (T, error) {
	var body T
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return body, fmt.Errorf("decoding request body: %w", err)
	}

	return body, nil
}

// errorResponse is the one shape every error is returned in.
type errorResponse struct {
	Error string       `json:"error"`
	Code  errcode.Code `json:"code"`
}

func (server *Server) respondWithJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		server.logger.Error("encoding response", "error", err)
	}
}

func (server *Server) respondWithError(w http.ResponseWriter, err error) {
	code := errcode.FromError(err)
	if code.Class() == errcode.ClassInternal {
		server.logger.Error("internal error", "error", err)
	}

	server.respondWithCode(w, code)
}

func (server *Server) respondWithCode(w http.ResponseWriter, code errcode.Code) {
	server.respondWithJSON(w, statusForClass(code.Class()), errorResponse{
		Error: code.Description(),
		Code:  code,
	})
}

func statusForClass(class errcode.Class) int {
	switch class {
	case errcode.ClassInvalid:
		return http.StatusBadRequest
	case errcode.ClassNotFound:
		return http.StatusNotFound
	case errcode.ClassConflict:
		return http.StatusConflict
	case errcode.ClassUnauthenticated:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
