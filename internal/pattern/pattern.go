package pattern

import (
	"math"
)

type (
	Pattern[T any] struct {
		Value         T
		BothDiagonals []BothDiagonals
		Grids         []Grid
		Explicit      []Explicit
	}

	BothDiagonals struct {
		X          int
		Y          int
		StartAt    int
		SkipCount  int
		MatchCount int
	}

	Grid struct {
		X      int
		Y      int
		Width  int
		Height int
	}

	Explicit struct {
		X int
		Y int
	}
)

type Group[T any] []Pattern[T]

func (group Group[T]) Get(x, y int) (T, bool) {
	for _, pattern := range group {
		if value, ok := pattern.Get(x, y); ok {
			return value, true
		}
	}

	return *new(T), false
}

func (pattern Pattern[T]) Get(x, y int) (T, bool) {
	if x == 0 && y == 0 {
		return pattern.Value, false
	}

	for _, e := range pattern.Explicit {
		if e.X == x && e.Y == y {
			return pattern.Value, true
		}
	}

	for _, d := range pattern.BothDiagonals {
		if d.match(x, y) {
			return pattern.Value, true
		}
	}

	for _, g := range pattern.Grids {
		if g.match(x, y) {
			return pattern.Value, true
		}
	}

	return *new(T), false
}

func (diagonals BothDiagonals) match(x, y int) bool {
	x = x - diagonals.X
	y = y - diagonals.Y

	if x != y && -x != y {
		return false
	}

	d := int(math.Abs(float64(x))) + diagonals.StartAt

	series := d % (diagonals.SkipCount + diagonals.MatchCount)
	return series < diagonals.MatchCount
}

func (grid Grid) match(x, y int) bool {
	x = x - grid.X
	y = y - grid.Y
	return x%(grid.Width-1) == 0 && y%(grid.Height-1) == 0
}
