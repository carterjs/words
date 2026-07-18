package api

import (
	"encoding/json"
	"net/http"

	"github.com/carterjs/words/internal/errcode"
	"github.com/carterjs/words/internal/words"
)

type (
	playerResponse struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Score int    `json:"score"`
	}

	challengeResponse struct {
		ChallengerID   string `json:"challengerId"`
		MoverID        string `json:"moverId"`
		VotesInvalid   int    `json:"votesInvalid"`
		VotesValid     int    `json:"votesValid"`
		VotesNeeded    int    `json:"votesNeeded"`
		EligibleVoters int    `json:"eligibleVoters"`
		Resolved       bool   `json:"resolved"`
		Upheld         bool   `json:"upheld"`
		RescindedWord  string `json:"rescindedWord,omitempty"`
	}

	gameResponse struct {
		ID                   string             `json:"id"`
		Started              bool               `json:"started"`
		Finished             bool               `json:"finished"`
		Round                int                `json:"round"`
		CurrentPlayerID      string             `json:"currentPlayerId"`
		LettersRemaining     int                `json:"lettersRemaining"`
		Players              []playerResponse   `json:"players"`
		LetterPoints         map[string]int     `json:"letterPoints"`
		WinnerIDs            []string           `json:"winnerIds,omitempty"`
		Challenge            *challengeResponse `json:"challenge,omitempty"`
		ChallengeableMoverID string             `json:"challengeableMoverId,omitempty"`
		PlayerID             string             `json:"playerId"`
		Rack                 []string           `json:"rack,omitempty"`
	}

	turnResponse struct {
		Round           int      `json:"round"`
		CurrentPlayerID string   `json:"currentPlayerId"`
		Finished        bool     `json:"finished"`
		WinnerIDs       []string `json:"winnerIds,omitempty"`
		Rack            []string `json:"rack,omitempty"`
	}
)

func (server *Server) handleCreateGame() http.HandlerFunc {
	type overridesBody struct {
		RackSize           int            `json:"rackSize,omitempty"`
		LetterDistribution map[string]int `json:"letterDistribution,omitempty"`
		LetterPoints       map[string]int `json:"letterPoints,omitempty"`
	}

	type requestBody struct {
		Preset    string        `json:"preset"`
		Overrides overridesBody `json:"overrides"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := parseRequestBody[requestBody](r)
		if err != nil {
			server.respondWithCode(w, errcode.BadRequest)
			return
		}

		game, err := server.service.CreateGame(r.Context(), body.Preset, words.ConfigOverrides{
			RackSize:           body.Overrides.RackSize,
			LetterDistribution: runeCounts(body.Overrides.LetterDistribution),
			LetterPoints:       runeCounts(body.Overrides.LetterPoints),
		})
		if err != nil {
			server.respondWithError(w, err)
			return
		}

		server.respondWithJSON(w, http.StatusCreated, constructGameResponse(r, game))
	}
}

func (server *Server) handleGetGameByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		game, err := server.service.GameByID(r.Context(), r.PathValue("gameId"))
		if err != nil {
			server.respondWithError(w, err)
			return
		}

		server.respondWithJSON(w, http.StatusOK, constructGameResponse(r, game))
	}
}

func (server *Server) handleUpdateGame() http.HandlerFunc {
	type requestBody struct {
		Operation string          `json:"operation"`
		Payload   json.RawMessage `json:"payload"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := parseRequestBody[requestBody](r)
		if err != nil {
			server.respondWithCode(w, errcode.BadRequest)
			return
		}

		gameID := r.PathValue("gameId")

		switch body.Operation {
		case "JOIN_GAME":
			server.joinGame(w, r, gameID, body.Payload)
		case "START_GAME":
			server.startGame(w, r, gameID)
		case "PASS_TURN":
			server.passTurn(w, r, gameID)
		case "EXCHANGE_LETTERS":
			server.exchangeLetters(w, r, gameID, body.Payload)
		case "CHALLENGE_WORD":
			server.challengeWord(w, r, gameID)
		case "CAST_VOTE":
			server.castVote(w, r, gameID, body.Payload)
		default:
			server.respondWithCode(w, errcode.UnknownOperation)
		}
	}
}

func (server *Server) joinGame(w http.ResponseWriter, r *http.Request, gameID string, payload json.RawMessage) {
	type joinResponse struct {
		PlayerID string           `json:"playerId"`
		Players  []playerResponse `json:"players"`
	}

	var request struct {
		PlayerName string `json:"playerName"`
	}
	if err := json.Unmarshal(payload, &request); err != nil {
		server.respondWithCode(w, errcode.BadRequest)
		return
	}

	game, player, err := server.service.JoinGame(r.Context(), gameID, request.PlayerName)
	if err != nil {
		server.respondWithError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     playerIDCookie,
		Value:    player.ID(),
		Path:     r.URL.Path,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
	})

	server.respondWithJSON(w, http.StatusCreated, joinResponse{
		PlayerID: player.ID(),
		Players:  constructPlayerResponses(game),
	})
}

func (server *Server) startGame(w http.ResponseWriter, r *http.Request, gameID string) {
	type startResponse struct {
		Started bool     `json:"started"`
		Rack    []string `json:"rack"`
	}

	game, err := server.service.StartGame(r.Context(), gameID)
	if err != nil {
		server.respondWithError(w, err)
		return
	}

	rack := []string{}
	if playerID, identified := playerIDFromRequest(r); identified {
		if player, exists := game.PlayerByID(playerID); exists {
			rack = letterStrings(player.Letters())
		}
	}

	server.respondWithJSON(w, http.StatusOK, startResponse{
		Started: game.Started(),
		Rack:    rack,
	})
}

func (server *Server) passTurn(w http.ResponseWriter, r *http.Request, gameID string) {
	playerID, identified := playerIDFromRequest(r)
	if !identified {
		server.respondWithCode(w, errcode.MissingPlayer)
		return
	}

	game, err := server.service.PassTurn(r.Context(), gameID, playerID)
	if err != nil {
		server.respondWithError(w, err)
		return
	}

	server.respondWithJSON(w, http.StatusOK, constructTurnResponse(game, playerID))
}

func (server *Server) exchangeLetters(w http.ResponseWriter, r *http.Request, gameID string, payload json.RawMessage) {
	playerID, identified := playerIDFromRequest(r)
	if !identified {
		server.respondWithCode(w, errcode.MissingPlayer)
		return
	}

	var request struct {
		Letters []string `json:"letters"`
	}
	if err := json.Unmarshal(payload, &request); err != nil {
		server.respondWithCode(w, errcode.BadRequest)
		return
	}

	letters, valid := lettersFromStrings(request.Letters)
	if !valid {
		server.respondWithCode(w, errcode.BadRequest)
		return
	}

	game, err := server.service.ExchangeLetters(r.Context(), gameID, playerID, letters)
	if err != nil {
		server.respondWithError(w, err)
		return
	}

	server.respondWithJSON(w, http.StatusOK, constructTurnResponse(game, playerID))
}

func (server *Server) challengeWord(w http.ResponseWriter, r *http.Request, gameID string) {
	playerID, identified := playerIDFromRequest(r)
	if !identified {
		server.respondWithCode(w, errcode.MissingPlayer)
		return
	}

	_, outcome, err := server.service.ChallengeWord(r.Context(), gameID, playerID)
	if err != nil {
		server.respondWithError(w, err)
		return
	}

	server.respondWithJSON(w, http.StatusOK, constructChallengeResponse(outcome))
}

func (server *Server) castVote(w http.ResponseWriter, r *http.Request, gameID string, payload json.RawMessage) {
	playerID, identified := playerIDFromRequest(r)
	if !identified {
		server.respondWithCode(w, errcode.MissingPlayer)
		return
	}

	var request struct {
		Vote words.Vote `json:"vote"`
	}
	if err := json.Unmarshal(payload, &request); err != nil {
		server.respondWithCode(w, errcode.BadRequest)
		return
	}

	_, outcome, err := server.service.CastVote(r.Context(), gameID, playerID, request.Vote)
	if err != nil {
		server.respondWithError(w, err)
		return
	}

	server.respondWithJSON(w, http.StatusOK, constructChallengeResponse(outcome))
}

func constructGameResponse(r *http.Request, game *words.Game) gameResponse {
	letterPoints := make(map[string]int)
	for letter, points := range game.Config().LetterPoints {
		letterPoints[string(letter)] = points
	}

	response := gameResponse{
		ID:               game.ID(),
		Started:          game.Started(),
		Finished:         game.Finished(),
		Round:            game.Round(),
		CurrentPlayerID:  game.CurrentPlayerID(),
		LettersRemaining: game.LettersRemaining(),
		Players:          constructPlayerResponses(game),
		LetterPoints:     letterPoints,
		WinnerIDs:        game.WinnerIDs(),
	}

	if outcome, pending := game.PendingChallenge(); pending {
		challenge := constructChallengeResponse(outcome)
		response.Challenge = &challenge
	}

	if moverID, challengeable := game.ChallengeableMoverID(); challengeable {
		response.ChallengeableMoverID = moverID
	}

	if playerID, identified := playerIDFromRequest(r); identified {
		if player, exists := game.PlayerByID(playerID); exists {
			response.PlayerID = playerID
			response.Rack = letterStrings(player.Letters())
		}
	}

	return response
}

func constructPlayerResponses(game *words.Game) []playerResponse {
	var players []playerResponse
	for _, player := range game.Players() {
		players = append(players, playerResponse{
			ID:    player.ID(),
			Name:  player.Name(),
			Score: player.Score(),
		})
	}

	return players
}

func constructTurnResponse(game *words.Game, playerID string) turnResponse {
	response := turnResponse{
		Round:           game.Round(),
		CurrentPlayerID: game.CurrentPlayerID(),
		Finished:        game.Finished(),
		WinnerIDs:       game.WinnerIDs(),
	}

	if player, exists := game.PlayerByID(playerID); exists {
		response.Rack = letterStrings(player.Letters())
	}

	return response
}

func constructChallengeResponse(outcome words.ChallengeOutcome) challengeResponse {
	response := challengeResponse{
		ChallengerID:   outcome.ChallengerID,
		MoverID:        outcome.MoverID,
		VotesInvalid:   outcome.VotesInvalid,
		VotesValid:     outcome.VotesValid,
		VotesNeeded:    outcome.VotesNeeded,
		EligibleVoters: outcome.EligibleVoters,
		Resolved:       outcome.Resolved,
		Upheld:         outcome.Upheld,
	}

	if outcome.RescindedWord != nil {
		response.RescindedWord = string(outcome.RescindedWord.Letters())
	}

	return response
}

func letterStrings(letters []rune) []string {
	strings := make([]string, len(letters))
	for index, letter := range letters {
		strings[index] = string(letter)
	}

	return strings
}

// runeCounts converts JSON letter keys to runes, skipping empty keys.
func runeCounts(counts map[string]int) map[rune]int {
	if len(counts) == 0 {
		return nil
	}

	runes := make(map[rune]int, len(counts))
	for letter, count := range counts {
		if letter == "" {
			continue
		}
		runes[[]rune(letter)[0]] = count
	}

	return runes
}

func lettersFromStrings(rawLetters []string) ([]rune, bool) {
	letters := make([]rune, 0, len(rawLetters))
	for _, rawLetter := range rawLetters {
		if rawLetter == "" {
			return nil, false
		}
		letters = append(letters, []rune(rawLetter)[0])
	}

	return letters, true
}
