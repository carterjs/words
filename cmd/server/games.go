package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/carterjs/words/internal/game"
)

func (server *Server) handleCreateGame() http.HandlerFunc {
	type request struct {
		Name       string `json:"name"`
		Passphrase string `json:"passphrase"`
	}

	type player struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Active bool   `json:"active"`
	}

	type response struct {
		ID      string   `json:"id"`
		Name    string   `json:"name"`
		Players []player `json:"players"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid request")
			return
		}

		game, err := server.Games.CreateGame(r.Context(), req.Name, req.Passphrase)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed to create game")
			return
		}

		respondWithJSON(w, http.StatusCreated, response{
			ID:   game.ID,
			Name: game.Name,
		})
	}
}

func (server *Server) handleGetGame() http.HandlerFunc {
	type response struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("gameId")

		game, err := server.Games.GetGameByID(r.Context(), id)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "game not found")
			return
		}

		respondWithJSON(w, http.StatusOK, response{
			ID:   game.ID,
			Name: game.Name,
		})
	}
}

func (server *Server) handleAddPlayerToGame() http.HandlerFunc {
	type request struct {
		Name       string `json:"name"`
		Passphrase string `json:"passphrase"`
	}

	type response struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		gameID := r.PathValue("gameId")

		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid request")
			return
		}

		p, err := server.Games.AddPlayerToGame(r.Context(), gameID, req.Name, req.Passphrase)
		if err != nil {
			if errors.Is(err, game.ErrIncorrectPassphrase) {
				respondWithError(w, http.StatusUnauthorized, "incorrect passphrase")
				return
			}

			respondWithError(w, http.StatusInternalServerError, "failed to add player to game")
			return
		}

		respondWithJSON(w, http.StatusCreated, response{
			ID:   p.ID,
			Name: p.Name,
		})
	}
}

func (server *Server) handleGetGamePlayers() http.HandlerFunc {
	type player struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		gameID := r.PathValue("gameId")

		players, err := server.Games.GetPlayersByGameID(r.Context(), gameID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "game not found")
			return
		}

		var respPlayers []player
		for _, p := range players {
			respPlayers = append(respPlayers, player{
				ID:     p.ID,
				Name:   p.Name,
				Status: string(p.Status),
			})
		}

		respondWithJSON(w, http.StatusOK, respPlayers)
	}
}
