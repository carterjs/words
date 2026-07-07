package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/carterjs/words/internal/api"
	"github.com/carterjs/words/internal/pubsub"
	"github.com/carterjs/words/internal/store"
	"github.com/carterjs/words/internal/words"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_Handler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{name: "reports an unknown game", method: http.MethodGet, path: "/api/v1/games/nope", wantStatus: http.StatusNotFound},
		{name: "rejects an unparsable body", method: http.MethodPost, path: "/api/v1/games", body: "{", wantStatus: http.StatusBadRequest},
		{name: "rejects an unknown operation", method: http.MethodPatch, path: "/api/v1/games/nope", body: `{"operation":"EXPLODE"}`, wantStatus: http.StatusBadRequest},
		{name: "requires a player to pass", method: http.MethodPatch, path: "/api/v1/games/nope", body: `{"operation":"PASS_TURN"}`, wantStatus: http.StatusUnauthorized},
		{name: "lists presets", method: http.MethodGet, path: "/api/v1/presets", wantStatus: http.StatusOK},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := newTestServer(t)

			request := httptest.NewRequest(test.method, test.path, bytes.NewBufferString(test.body))
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)

			assert.Equal(t, test.wantStatus, recorder.Code)
		})
	}
}

// TestServer_Handler_Integration drives a full two-player game through the
// HTTP API: create, join, start, play, and a successful challenge vote.
func TestServer_Handler_Integration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "plays a challenged word end to end"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			handler := newTestServer(t).Handler()
			client := &apiClient{t: t, handler: handler}

			created := client.do(http.MethodPost, "/api/v1/games", createGameBody(), "")
			gameID := created["id"].(string)
			gamePath := "/api/v1/games/" + gameID

			joinBody := `{"operation":"JOIN_GAME","payload":{"playerName":"one"}}`
			first := client.do(http.MethodPatch, gamePath, joinBody, "")
			firstID := first["playerId"].(string)

			joinBody = `{"operation":"JOIN_GAME","payload":{"playerName":"two"}}`
			second := client.do(http.MethodPatch, gamePath, joinBody, "")
			secondID := second["playerId"].(string)

			client.do(http.MethodPatch, gamePath, `{"operation":"START_GAME"}`, firstID)

			state := client.do(http.MethodGet, gamePath, "", firstID)
			mover := state["currentPlayerId"].(string)
			opponent := secondID
			if mover == secondID {
				opponent = firstID
			}

			playBody := `{"operation":"ADD_WORD","payload":{"x":0,"y":0,"direction":"HORIZONTAL","word":"AA"}}`
			played := client.do(http.MethodPatch, gamePath+"/board", playBody, mover)
			assert.Equal(t, float64(2), played["points"])

			challenge := client.do(http.MethodPatch, gamePath, `{"operation":"CHALLENGE_WORD"}`, opponent)
			assert.Equal(t, true, challenge["resolved"])
			assert.Equal(t, true, challenge["upheld"])

			// the rescinded word leaves no letters, only preset modifiers
			board := client.do(http.MethodGet, gamePath+"/board", "", "")
			for _, cell := range board["cells"].([]any) {
				assert.NotContains(t, cell.(map[string]any), "letter")
			}
		})
	}
}

// apiClient drives the handler with per-request player cookies.
type apiClient struct {
	t       *testing.T
	handler http.Handler
}

func (client *apiClient) do(method, path, body, playerID string) map[string]any {
	client.t.Helper()

	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if playerID != "" {
		request.AddCookie(&http.Cookie{Name: "playerId", Value: playerID})
	}

	recorder := httptest.NewRecorder()
	client.handler.ServeHTTP(recorder, request)
	require.Less(client.t, recorder.Code, http.StatusBadRequest, recorder.Body.String())

	var response map[string]any
	require.NoError(client.t, json.Unmarshal(recorder.Body.Bytes(), &response))

	return response
}

func newTestServer(t *testing.T) *api.Server {
	t.Helper()

	service := words.NewService(store.NewFS(t.TempDir()), pubsub.NewGameBroker(), slog.New(slog.DiscardHandler))

	return api.NewServer(service, slog.New(slog.DiscardHandler), api.Config{PublicDirectory: t.TempDir()})
}

// createGameBody zeroes out every standard letter except A so racks and
// plays are deterministic.
func createGameBody() string {
	distribution := map[string]int{"A": 20, "_": 0}
	for letter := 'B'; letter <= 'Z'; letter++ {
		distribution[string(letter)] = 0
	}

	encoded, err := json.Marshal(distribution)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf(`{"preset":"standard","overrides":{"rackSize":3,"letterDistribution":%s}}`, encoded)
}
