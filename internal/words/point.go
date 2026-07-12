package words

import (
	"fmt"
	"strconv"
	"strings"
)

// Point identifies a single cell on the unbounded board, encoded as
// "column,row" so it can serve as a JSON map key.
type Point string

// NewPoint returns the point at the given column and row.
func NewPoint(column, row int) Point {
	return Point(fmt.Sprint(column, ",", row))
}

// Column returns the horizontal coordinate of the point.
func (point Point) Column() int {
	rawColumn, _, _ := strings.Cut(string(point), ",")
	column, _ := strconv.Atoi(rawColumn)
	return column
}

// Row returns the vertical coordinate of the point.
func (point Point) Row() int {
	_, rawRow, _ := strings.Cut(string(point), ",")
	row, _ := strconv.Atoi(rawRow)
	return row
}

// Offset returns the point moved by the given column and row deltas.
func (point Point) Offset(deltaColumn, deltaRow int) Point {
	return NewPoint(point.Column()+deltaColumn, point.Row()+deltaRow)
}
