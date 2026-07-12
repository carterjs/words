package api

import (
	"net/http"

	"github.com/carterjs/words/internal/words"
)

type presetResponse struct {
	ID                 string         `json:"id"`
	Name               string         `json:"name"`
	Description        string         `json:"description"`
	RackSize           int            `json:"rackSize"`
	LetterDistribution map[string]int `json:"letterDistribution"`
	LetterPoints       map[string]int `json:"letterPoints"`
}

func (server *Server) handleGetPresets() http.HandlerFunc {
	type responseBody struct {
		Presets []presetResponse `json:"presets"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var presets []presetResponse
		for _, preset := range words.Presets {
			presets = append(presets, constructPresetResponse(preset))
		}

		server.respondWithJSON(w, http.StatusOK, responseBody{Presets: presets})
	}
}

func (server *Server) handleGetPresetByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		preset, exists := words.PresetByID(r.PathValue("id"))
		if !exists {
			server.respondWithError(w, words.ErrPresetNotFound)
			return
		}

		server.respondWithJSON(w, http.StatusOK, constructPresetResponse(preset))
	}
}

func (server *Server) handleGetPresetBoard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		preset, exists := words.PresetByID(r.PathValue("id"))
		if !exists {
			server.respondWithError(w, words.ErrPresetNotFound)
			return
		}

		area := parseExtents(r, extents{
			minX: -defaultBoardExtent,
			minY: -defaultBoardExtent,
			maxX: defaultBoardExtent,
			maxY: defaultBoardExtent,
		})

		var cells []cellResponse
		for row := area.minY; row <= area.maxY; row++ {
			for column := area.minX; column <= area.maxX; column++ {
				if modifier, hasModifier := preset.Modifiers.Get(column, row); hasModifier {
					cells = append(cells, cellResponse{
						X:        column,
						Y:        row,
						Modifier: string(modifier),
					})
				}
			}
		}

		server.respondWithJSON(w, http.StatusOK, boardResponse{Cells: cells})
	}
}

func constructPresetResponse(preset words.Preset) presetResponse {
	response := presetResponse{
		ID:                 preset.ID,
		Name:               preset.Name,
		Description:        preset.Description,
		RackSize:           preset.RackSize,
		LetterDistribution: make(map[string]int),
		LetterPoints:       make(map[string]int),
	}

	for letter, count := range preset.LetterDistribution {
		response.LetterDistribution[string(letter)] = count
	}

	for letter, points := range preset.LetterPoints {
		response.LetterPoints[string(letter)] = points
	}

	return response
}
