package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// keepaliveInterval is how often a comment is written to an otherwise quiet
// event stream so proxies don't drop the connection as idle.
const keepaliveInterval = 25 * time.Second

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

		// flush the headers right away so clients see the stream as open and
		// can catch up on state missed while (re)connecting
		if canFlush {
			flusher.Flush()
		}

		for {
			waitCtx, cancel := context.WithTimeout(r.Context(), keepaliveInterval)
			event, err := subscription.Next(waitCtx)
			cancel()

			if err != nil {
				if r.Context().Err() != nil || !errors.Is(err, context.DeadlineExceeded) {
					return
				}

				// quiet stretch: write a comment so the connection stays alive
				fmt.Fprint(w, ": keepalive\n\n")
				if canFlush {
					flusher.Flush()
				}
				continue
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
