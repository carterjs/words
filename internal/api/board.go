package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/carterjs/words/internal/errcode"
	"github.com/carterjs/words/internal/words"
)

// defaultBoardExtent is how far from the center the board is reported when
// the caller does not ask for a specific window.
const defaultBoardExtent = 15

type (
	cellResponse struct {
		X        int    `json:"x"`
		Y        int    `json:"y"`
		Letter   string `json:"letter,omitempty"`
		Modifier string `json:"modifier,omitempty"`
	}

	boardResponse struct {
		Cells []cellResponse `json:"cells"`
	}

	// extents is an inclusive window of board cells.
	extents struct {
		minX, minY, maxX, maxY int
	}

	placementResponse struct {
		X             int            `json:"x"`
		Y             int            `json:"y"`
		Direction     string         `json:"direction"`
		Word          string         `json:"word"`
		Points        int            `json:"points"`
		IndirectWords []indirectWord `json:"indirectWords"`
	}

	indirectWord struct {
		X         int    `json:"x"`
		Y         int    `json:"y"`
		Direction string `json:"direction"`
		Word      string `json:"word"`
	}
)

func (server *Server) handleGetGameBoard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		game, err := server.service.GameByID(r.Context(), r.PathValue("gameId"))
		if err != nil {
			server.respondWithError(w, err)
			return
		}

		area := parseExtents(r, extentsCovering(game.Board().Bounds()))

		server.respondWithJSON(w, http.StatusOK, boardResponse{
			Cells: boardCells(game.Board(), area),
		})
	}
}

func (server *Server) handleUpdateBoard() http.HandlerFunc {
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

		switch body.Operation {
		case "ADD_WORD":
			server.addWordToBoard(w, r, r.PathValue("gameId"), body.Payload)
		default:
			server.respondWithCode(w, errcode.UnknownOperation)
		}
	}
}

func (server *Server) addWordToBoard(w http.ResponseWriter, r *http.Request, gameID string, payload json.RawMessage) {
	playerID, identified := playerIDFromRequest(r)
	if !identified {
		server.respondWithCode(w, errcode.MissingPlayer)
		return
	}

	var request struct {
		X         int    `json:"x"`
		Y         int    `json:"y"`
		Direction string `json:"direction"`
		Word      string `json:"word"`
	}
	if err := json.Unmarshal(payload, &request); err != nil {
		server.respondWithCode(w, errcode.BadRequest)
		return
	}

	direction := words.Direction(request.Direction)
	if direction != words.DirectionHorizontal && direction != words.DirectionVertical {
		server.respondWithCode(w, errcode.BadRequest)
		return
	}

	word := words.NewWord(words.NewPoint(request.X, request.Y), direction, request.Word)

	_, result, err := server.service.PlayWord(r.Context(), gameID, playerID, word)
	if err != nil {
		server.respondWithError(w, err)
		return
	}

	server.respondWithJSON(w, http.StatusOK, constructPlacementResponse(result))
}

func (server *Server) handleGetGameBoardPlacements() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playerID, identified := playerIDFromRequest(r)
		if !identified {
			server.respondWithCode(w, errcode.MissingPlayer)
			return
		}

		column, columnErr := strconv.Atoi(r.URL.Query().Get("x"))
		row, rowErr := strconv.Atoi(r.URL.Query().Get("y"))
		word := r.URL.Query().Get("word")
		if columnErr != nil || rowErr != nil || word == "" {
			server.respondWithCode(w, errcode.BadRequest)
			return
		}

		game, err := server.service.GameByID(r.Context(), r.PathValue("gameId"))
		if err != nil {
			server.respondWithError(w, err)
			return
		}

		placements, err := game.FindPlacements(playerID, words.NewPoint(column, row), word)
		if err != nil {
			server.respondWithError(w, err)
			return
		}

		responses := make([]placementResponse, 0, len(placements))
		for _, placement := range placements {
			responses = append(responses, constructPlacementResponse(placement))
		}

		server.respondWithJSON(w, http.StatusOK, responses)
	}
}

func constructPlacementResponse(result words.PlacementResult) placementResponse {
	start := result.DirectWord.Start()
	response := placementResponse{
		X:         start.Column(),
		Y:         start.Row(),
		Direction: string(result.DirectWord.Direction()),
		Word:      string(result.DirectWord.Letters()),
		Points:    result.Points,
	}

	for _, indirect := range result.IndirectWords {
		indirectStart := indirect.Start()
		response.IndirectWords = append(response.IndirectWords, indirectWord{
			X:         indirectStart.Column(),
			Y:         indirectStart.Row(),
			Direction: string(indirect.Direction()),
			Word:      string(indirect.Letters()),
		})
	}

	return response
}

func boardCells(board *words.Board, area extents) []cellResponse {
	var cells []cellResponse

	for row := area.minY; row <= area.maxY; row++ {
		for column := area.minX; column <= area.maxX; column++ {
			cell := cellResponse{X: column, Y: row}

			if letter, occupied := board.Letter(words.NewPoint(column, row)); occupied {
				cell.Letter = string(letter)
			}

			if modifier, hasModifier := board.Modifier(words.NewPoint(column, row)); hasModifier {
				cell.Modifier = string(modifier)
			}

			if cell.Letter != "" || cell.Modifier != "" {
				cells = append(cells, cell)
			}
		}
	}

	return cells
}

// extentsCovering widens the default window to cover everything placed.
func extentsCovering(bounds words.Bounds) extents {
	return extents{
		minX: min(-defaultBoardExtent, bounds.MinX),
		minY: min(-defaultBoardExtent, bounds.MinY),
		maxX: max(defaultBoardExtent, bounds.MaxX),
		maxY: max(defaultBoardExtent, bounds.MaxY),
	}
}

func parseExtents(r *http.Request, defaults extents) extents {
	return extents{
		minX: queryInt(r, "minX", defaults.minX),
		minY: queryInt(r, "minY", defaults.minY),
		maxX: queryInt(r, "maxX", defaults.maxX),
		maxY: queryInt(r, "maxY", defaults.maxY),
	}
}

func queryInt(r *http.Request, key string, fallback int) int {
	value, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil {
		return fallback
	}

	return value
}
