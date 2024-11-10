package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/carterjs/words/internal/words"
	"net/http"
	"strconv"
	"strings"
)

type (
	gameResponse struct {
		ID       string           `json:"id"`
		Started  bool             `json:"started"`
		Players  []playerResponse `json:"players"`
		PlayerID string           `json:"playerId"`
		Rack     []string         `json:"rack"`
	}

	partialGameResponse struct {
		ID       string           `json:"id,omitempty"`
		Started  bool             `json:"started,omitempty"`
		Players  []playerResponse `json:"players,omitempty"`
		PlayerID string           `json:"playerId,omitempty"`
		Rack     []string         `json:"rack,omitempty"`
	}

	playerResponse struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Score int    `json:"score"`
	}

	cellResponse struct {
		X        int    `json:"x"`
		Y        int    `json:"y"`
		Letter   string `json:"letter,omitempty"`
		Modifier string `json:"modifier,omitempty"`
	}
)

func (server *Server) handleCreateGame() http.HandlerFunc {
	type (
		presetOverrides struct {
			RackSize           int            `json:"rackSize,omitempty"`
			LetterDistribution map[string]int `json:"letterDistribution,omitempty"`
			LetterPoints       map[string]int `json:"letterPoints,omitempty"`
		}

		requestBody struct {
			Preset          string          `json:"preset"`
			PresetOverrides presetOverrides `json:"overrides"`
		}
	)

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := parseRequestBody[requestBody](r)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		preset := getPresetByID(body.Preset)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "failed to get preset")
			return
		}

		config := preset.Config

		if body.PresetOverrides.RackSize != 0 {
			config.RackSize = body.PresetOverrides.RackSize
		}

		if len(body.PresetOverrides.LetterDistribution) > 0 {
			config.LetterDistribution = preset.Config.LetterDistribution
			for k, v := range body.PresetOverrides.LetterDistribution {
				config.LetterDistribution[rune(k[0])] = v
			}
		}

		if len(body.PresetOverrides.LetterPoints) > 0 {
			config.LetterPoints = preset.Config.LetterPoints
			for k, v := range body.PresetOverrides.LetterPoints {
				config.LetterPoints[rune(k[0])] = v
			}
		}

		game := words.NewGame(config)
		if err := server.gameStore.SaveGame(r.Context(), game); err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed to save game")
			return
		}

		respondWithJSON(w, http.StatusCreated, constructGameResponse(r, game))
	}
}

func constructGameResponse(r *http.Request, game *words.Game) gameResponse {
	players := make([]playerResponse, len(game.Players))
	for i, p := range game.Players {
		players[i] = constructPlayerResponse(&p)
	}

	minX := -15
	minY := -15
	maxX := 15
	maxY := 15

	if game.Board.MinX < minX {
		minX = game.Board.MinX
	}
	if game.Board.MinY < minY {
		minY = game.Board.MinY
	}
	if game.Board.MaxX > maxX {
		maxX = game.Board.MaxX
	}
	if game.Board.MaxY > maxY {
		maxY = game.Board.MaxY
	}

	resp := gameResponse{
		ID:      game.ID,
		Started: game.Started,
		Players: players,
	}

	playerID, exists := getPlayerID(r)
	if exists {
		resp.PlayerID = playerID
		resp.Rack = make([]string, len(game.GetPlayerByID(playerID).Letters))
		for i, l := range game.GetPlayerByID(playerID).Letters {
			resp.Rack[i] = string(l)
		}
	}

	return resp
}

func constructPlayerResponse(player *words.Player) playerResponse {
	return playerResponse{
		ID:    player.ID,
		Name:  player.Name,
		Score: player.Score(),
	}
}

// TODO: buffer border
func constructBoardResponse(board *words.Board, minX int, minY int, maxX int, maxY int) []cellResponse {
	var cells []cellResponse

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			cell := cellResponse{
				X: x,
				Y: y,
			}

			if letter, hasLetter := board.GetLetter(words.NewPoint(x, y)); hasLetter {
				cell.Letter = string(letter)
			}

			if modifier, hasModifier := board.GetModifier(words.NewPoint(x, y)); hasModifier {
				cell.Modifier = string(modifier)
			}

			if cell.Letter != "" || cell.Modifier != "" {
				cells = append(cells, cell)
			}
		}
	}

	return cells
}

func (server *Server) handleGetGameByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("gameId")
		game, err := server.gameStore.GetGameByID(r.Context(), id)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "game not found")
			return
		}

		if game == nil {
			respondWithError(w, http.StatusNotFound, "game not found")
			return
		}

		respondWithJSON(w, http.StatusOK, constructGameResponse(r, game))
	}
}

type JoinGameRequest struct {
	PlayerName string `json:"playerName"`
}

func (server *Server) handleUpdateGame() http.HandlerFunc {
	type requestBody struct {
		Operation string          `json:"operation"`
		Payload   json.RawMessage `json:"payload"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		gameID := r.PathValue("gameId")

		body, err := parseRequestBody[requestBody](r)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		game, err := server.gameStore.GetGameByID(r.Context(), gameID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "game not found")
			return
		}

		switch body.Operation {
		case "START_GAME":
			server.startGame(w, r, game)
		case "JOIN_GAME":
			var joinGameRequest JoinGameRequest
			if err := json.Unmarshal(body.Payload, &joinGameRequest); err != nil {
				respondWithError(w, http.StatusBadRequest, "failed to unmarshal payload")
				return
			}

			server.addPlayerToGame(w, r, game, joinGameRequest)
		}
	}
}

func (server *Server) startGame(w http.ResponseWriter, r *http.Request, game *words.Game) {
	err := game.Start()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to start game: "+err.Error())
		return
	}

	// SAVE
	if err := server.gameStore.SaveGame(r.Context(), game); err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to save game")
		return
	}

	// for each player
	for _, p := range game.Players {
		var letters []string
		for _, l := range p.Letters {
			letters = append(letters, string(l))
		}

		server.events.Publish(r.Context(), gamePlayerChannel(game.ID, p.ID), GameStartedEvent{
			Letters: letters,
		})
	}

	var rack []string
	for _, l := range game.GetPlayerByID(game.Players[0].ID).Letters {
		rack = append(rack, string(l))
	}

	respondWithJSON(w, http.StatusOK, partialGameResponse{
		Started: true,
		Rack:    rack,
	})
}

func (server *Server) addPlayerToGame(w http.ResponseWriter, r *http.Request, game *words.Game, request JoinGameRequest) {
	player, err := game.AddPlayer(request.PlayerName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to add player: "+err.Error())
		return
	}

	if err := server.gameStore.SaveGame(r.Context(), game); err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to save game")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "playerId",
		Value: player.ID,
		// set cookie for the game .../games/{id}
		Path:     strings.TrimSuffix(r.URL.Path, "/players"),
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
	})

	var players []playerResponse
	for _, p := range game.Players {
		players = append(players, constructPlayerResponse(&p))
	}

	respondWithJSON(w, http.StatusCreated, partialGameResponse{
		PlayerID: player.ID,
		Players:  players,
	})
}

type boardResponse struct {
	Cells []cellResponse `json:"cells"`
}

func (server *Server) handleGetGameBoard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameID := r.PathValue("gameId")

		game, err := server.gameStore.GetGameByID(r.Context(), gameID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "game not found")
			return
		}

		minX := -15
		minY := -15
		maxX := 15
		maxY := 15

		if game.Board.MinX < minX {
			minX = game.Board.MinX
		}
		if game.Board.MinY < minY {
			minY = game.Board.MinY
		}
		if game.Board.MaxX > maxX {
			maxX = game.Board.MaxX
		}
		if game.Board.MaxY > maxY {
			maxY = game.Board.MaxY
		}

		minX, minY, maxX, maxY = parseBoardExtentsWithDefault(r, minX, minY, maxX, maxY)

		var cells []cellResponse

		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				cell := cellResponse{
					X: x,
					Y: y,
				}

				if letter, hasLetter := game.Board.GetLetter(words.NewPoint(x, y)); hasLetter {
					cell.Letter = string(letter)
				}

				if modifier, hasModifier := game.Board.GetModifier(words.NewPoint(x, y)); hasModifier {
					cell.Modifier = string(modifier)
				}

				if cell.Letter != "" || cell.Modifier != "" {
					cells = append(cells, cell)
				}
			}
		}

		respondWithJSON(w, http.StatusOK, boardResponse{Cells: cells})
	}
}

func parseBoardExtentsWithDefault(r *http.Request, defaultMinX, defaultMinY, defaultMaxX, defaultMaxY int) (int, int, int, int) {
	minX, err := strconv.Atoi(r.URL.Query().Get("minX"))
	if err != nil {
		minX = defaultMinX
	}

	minY, err := strconv.Atoi(r.URL.Query().Get("minY"))
	if err != nil {
		minY = defaultMinY
	}

	maxX, err := strconv.Atoi(r.URL.Query().Get("maxX"))
	if err != nil {
		maxX = defaultMaxX
	}

	maxY, err := strconv.Atoi(r.URL.Query().Get("maxY"))
	if err != nil {
		maxY = defaultMaxY
	}

	return minX, minY, maxX, maxY
}

type placementRequest struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Direction string `json:"direction"`
	Word      string `json:"word"`
}

type placementResponse struct {
	X             int            `json:"x"`
	Y             int            `json:"y"`
	Direction     string         `json:"direction"`
	Word          string         `json:"word"`
	Points        int            `json:"points"`
	IndirectWords []indirectWord `json:"indirectWords"`
}

type indirectWord struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Direction string `json:"direction"`
	Word      string `json:"word"`
}

func (server *Server) handleUpdateBoard() http.HandlerFunc {
	type requestBody struct {
		Operation string          `json:"operation"`
		Payload   json.RawMessage `json:"payload"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		gameID := r.PathValue("gameId")
		playerID, exists := getPlayerID(r)
		if !exists {
			respondWithError(w, http.StatusUnauthorized, "player not found")
			return
		}

		body, err := parseRequestBody[requestBody](r)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		game, err := server.gameStore.GetGameByID(r.Context(), gameID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "game not found")
			return
		}

		player := game.GetPlayerByID(playerID)
		if player == nil {
			respondWithError(w, http.StatusNotFound, "player not found")
			return
		}

		switch body.Operation {
		case "ADD_WORD":
			var request placementRequest
			if err := json.Unmarshal(body.Payload, &request); err != nil {
				respondWithError(w, http.StatusBadRequest, "failed to unmarshal payload")
				return
			}

			server.addWordToBoard(r.Context(), w, game, player.ID, request)
		}
	}
}

func (server *Server) addWordToBoard(ctx context.Context, w http.ResponseWriter, game *words.Game, playerID string, body placementRequest) {
	word := words.NewWord(words.NewPoint(body.X, body.Y), words.Direction(body.Direction), body.Word)
	placementResult, err := game.PlayWord(playerID, word)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := server.gameStore.SaveGame(ctx, game); err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to save game")
		return
	}

	respondWithJSON(w, http.StatusOK, constructPlacementResponse(placementResult))
}

func constructPlacementResponse(result words.PlacementResult) placementResponse {
	response := placementResponse{
		X:         result.DirectWord.Start.X(),
		Y:         result.DirectWord.Start.Y(),
		Direction: string(result.DirectWord.Direction),
		Word:      string(result.DirectWord.Letters),
		Points:    result.Points,
	}

	for _, i := range result.IndirectWords {
		response.IndirectWords = append(response.IndirectWords, indirectWord{
			X:         i.Start.X(),
			Y:         i.Start.Y(),
			Direction: string(i.Direction),
			Word:      string(i.Letters),
		})
	}

	return response
}

func (server *Server) handleGetGameBoardPlacements() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameID := r.PathValue("gameId")
		playerID, exists := getPlayerID(r)
		if !exists {
			respondWithError(w, http.StatusUnauthorized, "player not found")
			return
		}

		game, err := server.gameStore.GetGameByID(r.Context(), gameID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "game not found")
			return
		}

		player := game.GetPlayerByID(playerID)
		if player == nil {
			respondWithError(w, http.StatusNotFound, "player not found")
			return
		}

		x, err := strconv.Atoi(r.URL.Query().Get("x"))
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid x")
			return
		}

		y, err := strconv.Atoi(r.URL.Query().Get("y"))
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid y")
			return
		}

		word := r.URL.Query().Get("word")
		if word == "" {
			respondWithError(w, http.StatusBadRequest, "missing word")
			return
		}

		placements, err := game.FindPlacements(playerID, words.NewPoint(x, y), word)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		var placementResponses []placementResponse
		for _, p := range placements {
			placementResponses = append(placementResponses, constructPlacementResponse(p))
		}

		respondWithJSON(w, http.StatusOK, placementResponses)
	}
}

func (server *Server) handleGetPlayerRack() http.HandlerFunc {
	type responseBody struct {
		Letters []string `json:"letters"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		gameID := r.PathValue("gameId")
		playerID, exists := getPlayerID(r)
		if !exists {
			respondWithError(w, http.StatusUnauthorized, "player not found")
			return
		}

		game, err := server.gameStore.GetGameByID(r.Context(), gameID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "game not found")
			return
		}

		player := game.GetPlayerByID(playerID)
		if player == nil {
			respondWithError(w, http.StatusNotFound, "player not found")
			return
		}

		var letters []string
		for _, l := range player.Letters {
			letters = append(letters, string(l))
		}

		respondWithJSON(w, http.StatusOK, responseBody{Letters: letters})
	}
}

func (server *Server) handleStreamGameEvents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// server sent events
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		gameID := r.PathValue("gameId")
		playerID, _ := getPlayerID(r)

		// subscribe to game events
		events, unsubscribe := server.events.Subscribe(r.Context(), gameChannel(gameID), gamePlayerChannel(gameID, playerID))
		defer unsubscribe()

		for {
			select {
			case <-r.Context().Done():
				return
			case event := <-events:
				// steam event to user
				fmt.Fprintf(w, "event: %s\n", event.Type())

				b, err := json.Marshal(event)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				fmt.Fprintf(w, "data: %s\n\n", string(b))

				// flush response
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}

			}
		}
	}
}
