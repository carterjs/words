package pattern_test

import (
	"testing"

	"github.com/carterjs/words/internal/pattern"
	"github.com/stretchr/testify/assert"
)

func TestGroup_Get(t *testing.T) {
	t.Parallel()

	grids := pattern.Group[string]{
		{Value: "grid", Grids: []pattern.Grid{{Width: 5, Height: 5}}},
	}
	offsetGrid := pattern.Group[string]{
		{Value: "grid", Grids: []pattern.Grid{{X: 2, Y: 2, Width: 5, Height: 5}}},
	}
	diagonals := pattern.Group[string]{
		{Value: "diagonal", BothDiagonals: []pattern.BothDiagonals{{StartAt: 3, SkipCount: 2, MatchCount: 4}}},
	}
	layered := pattern.Group[string]{
		{Value: "first", Grids: []pattern.Grid{{Width: 5, Height: 5}}},
		{Value: "second", Grids: []pattern.Grid{{Width: 3, Height: 3}}},
	}

	tests := []struct {
		name      string
		group     pattern.Group[string]
		column    int
		row       int
		want      string
		wantMatch bool
	}{
		{name: "matches a grid at its intervals", group: grids, column: 4, row: 4, want: "grid", wantMatch: true},
		{name: "matches a grid across the axis", group: grids, column: -4, row: 8, want: "grid", wantMatch: true},
		{name: "misses between grid intervals", group: grids, column: 2, row: 4},
		{name: "never matches the center", group: grids, column: 0, row: 0},
		{name: "matches a grid offset from center", group: offsetGrid, column: 6, row: 6, want: "grid", wantMatch: true},
		{name: "matches a diagonal series cell", group: diagonals, column: 3, row: 3, want: "diagonal", wantMatch: true},
		{name: "matches the anti-diagonal", group: diagonals, column: -3, row: 3, want: "diagonal", wantMatch: true},
		{name: "misses a skipped diagonal cell", group: diagonals, column: 1, row: 1},
		{name: "prefers the first matching rule", group: layered, column: 4, row: 4, want: "first", wantMatch: true},
		{name: "falls through to later rules", group: layered, column: 2, row: 2, want: "second", wantMatch: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			value, matched := test.group.Get(test.column, test.row)

			assert.Equal(t, test.wantMatch, matched)
			assert.Equal(t, test.want, value)
		})
	}
}
