package main

import (
	"github.com/carterjs/words/internal/words"
	"net/http"
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
		for _, p := range words.Presets {
			presets = append(presets, constructPresetResponse(p))
		}

		respondWithJSON(w, http.StatusOK, responseBody{Presets: presets})
	}
}

func constructPresetResponse(preset words.Preset) presetResponse {
	response := presetResponse{
		ID:          preset.ID,
		Name:        preset.Name,
		Description: preset.Description,
		RackSize:    preset.RackSize,
	}

	response.LetterDistribution = make(map[string]int)
	for k, v := range preset.LetterDistribution {
		response.LetterDistribution[string(k)] = v
	}

	response.LetterPoints = make(map[string]int)
	for k, v := range preset.LetterPoints {
		response.LetterPoints[string(k)] = v
	}

	return response
}

func (server *Server) handleGetPresetByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		preset := getPresetByID(id)
		if preset == nil {
			respondWithError(w, http.StatusNotFound, "preset not found")
			return
		}

		respondWithJSON(w, http.StatusOK, constructPresetResponse(*preset))
	}
}

func getPresetByID(id string) *words.Preset {
	for _, p := range words.Presets {
		if p.ID == id {
			return &p
		}
	}

	return nil
}

func (server *Server) handleGetPresetBoard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		preset := getPresetByID(id)
		if preset == nil {
			respondWithError(w, http.StatusNotFound, "preset not found")
			return
		}

		minX, minY, maxX, maxY := parseBoardExtentsWithDefault(r, -15, -15, 15, 15)

		var cells []cellResponse

		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				if modifier, hasModifier := preset.Config.Modifiers.Get(x, y); hasModifier {
					cells = append(cells, cellResponse{
						X:        x,
						Y:        y,
						Modifier: string(modifier),
					})
				}
			}
		}

		respondWithJSON(w, http.StatusOK, boardResponse{
			Cells: cells,
		})
	}
}
