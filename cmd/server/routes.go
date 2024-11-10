package main

import (
	"net/http"
	"strings"
)

func (server *Server) Handler() http.HandlerFunc {
	mux := http.NewServeMux()

	// get presets
	mux.Handle("GET /api/v1/presets", server.handleGetPresets())
	mux.Handle("GET /api/v1/presets/{id}", server.handleGetPresetByID())
	mux.Handle("GET /api/v1/presets/{id}/board", server.handleGetPresetBoard())

	// create game
	mux.Handle("POST /api/v1/games", server.handleCreateGame())

	// get and update game
	mux.Handle("GET /api/v1/games/{gameId}", server.handleGetGameByID())
	mux.Handle("PATCH /api/v1/games/{gameId}", server.handleUpdateGame())

	// board
	mux.Handle("GET /api/v1/games/{gameId}/board", server.handleGetGameBoard())
	mux.Handle("GET /api/v1/games/{gameId}/board/placements", server.handleGetGameBoardPlacements())
	mux.Handle("PATCH /api/v1/games/{gameId}/board", server.handleUpdateBoard())

	// stream game events
	mux.Handle("GET /api/v1/games/{gameId}/events", server.handleStreamGameEvents())

	// fallback to public routes
	// TODO: 404 gracefully
	mux.Handle("/", handlePublic(publicDir))

	return withCORS(mux, "GET", "POST", "PATCH")
}

func withCORS(h http.Handler, methods ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h.ServeHTTP(w, r)
	}
}

func getPlayerID(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("playerId")
	if err == nil {
		return cookie.Value, true
	}

	return "", false
}

func handlePublic(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))

	return fs
}
