package game_test

import (
	"testing"

	"github.com/carterjs/words/internal/game"
)

func TestWordStats(t *testing.T) {
	stats := game.WordStats{
		Usages:     100,
		Approvals:  99,
		Rejections: 1,
	}

	t.Log(stats.Reputation())
}
