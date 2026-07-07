// Package pattern describes repeating geometric placements of values on an
// unbounded grid, used to lay out board modifiers without enumerating cells.
package pattern

import (
	"math"
)

type (
	// Rule pairs a value with the geometric shapes that place it on the grid.
	Rule[T any] struct {
		Value         T               `json:"value"`
		BothDiagonals []BothDiagonals `json:"bothDiagonals"`
		Grids         []Grid          `json:"grids"`
	}

	// BothDiagonals places values along the two diagonals crossing at a
	// center point, in a repeating series of matched and skipped cells.
	BothDiagonals struct {
		X          int `json:"x"`
		Y          int `json:"y"`
		StartAt    int `json:"startAt"`
		SkipCount  int `json:"skipCount"`
		MatchCount int `json:"matchCount"`
	}

	// Grid places values at regular column and row intervals across the
	// whole plane, offset from a center point.
	Grid struct {
		X      int `json:"x"`
		Y      int `json:"y"`
		Width  int `json:"width"`
		Height int `json:"height"`
	}
)

// Group is an ordered collection of rules; the first matching rule wins.
type Group[T any] []Rule[T]

// Get returns the value placed at the given coordinates and whether any rule
// in the group places one there. The origin is never matched: it is the
// board's center cell.
func (group Group[T]) Get(column, row int) (T, bool) {
	for _, rule := range group {
		if value, matched := ruleValueAt(rule, column, row); matched {
			return value, true
		}
	}

	return *new(T), false
}

func ruleValueAt[T any](rule Rule[T], column, row int) (T, bool) {
	if column == 0 && row == 0 {
		return *new(T), false
	}

	for _, diagonals := range rule.BothDiagonals {
		if matchDiagonals(diagonals, column, row) {
			return rule.Value, true
		}
	}

	for _, grid := range rule.Grids {
		if matchGrid(grid, column, row) {
			return rule.Value, true
		}
	}

	return *new(T), false
}

func matchDiagonals(diagonals BothDiagonals, column, row int) bool {
	offsetX := column - diagonals.X
	offsetY := row - diagonals.Y

	if offsetX != offsetY && -offsetX != offsetY {
		return false
	}

	if diagonals.SkipCount == 0 && diagonals.MatchCount == 0 {
		return false
	}

	distance := int(math.Abs(float64(offsetX))) + diagonals.StartAt
	series := distance % (diagonals.SkipCount + diagonals.MatchCount)

	return series < diagonals.MatchCount
}

func matchGrid(grid Grid, column, row int) bool {
	if grid.Width <= 1 || grid.Height <= 1 {
		return false
	}

	offsetX := column - grid.X
	offsetY := row - grid.Y

	return offsetX%(grid.Width-1) == 0 && offsetY%(grid.Height-1) == 0
}
