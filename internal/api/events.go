package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// handleStreamGameEvents streams a game's events over server-sent events.
// Players identified by their cookie also receive their private events.
func (server *Server) handleStreamGameEvents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		gameID := r.PathValue("gameId")
		playerID, _ := playerIDFromRequest(r)

		subscription := server.service.Subscribe(r.Context(), gameID, playerID)
		defer subscription.Close()

		flusher, canFlush := w.(http.Flusher)

		for {
			event, err := subscription.Next(r.Context())
			if err != nil {
				return
			}

			payload := event.Payload
			if payload == nil {
				payload = json.RawMessage("{}")
			}

			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, payload)

			if canFlush {
				flusher.Flush()
			}
		}
	}
}
