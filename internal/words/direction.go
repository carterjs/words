package words

// Direction is the orientation in which a word is laid on the board.
type Direction string

const (
	// DirectionHorizontal lays letters left to right.
	DirectionHorizontal Direction = "HORIZONTAL"
	// DirectionVertical lays letters top to bottom.
	DirectionVertical Direction = "VERTICAL"
)

// Vector returns the column and row deltas for moving the given number of
// steps along the direction.
func (direction Direction) Vector(steps int) (int, int) {
	if direction == DirectionHorizontal {
		return steps, 0
	}

	return 0, steps
}

// Other returns the perpendicular direction.
func (direction Direction) Other() Direction {
	if direction == DirectionHorizontal {
		return DirectionVertical
	}

	return DirectionHorizontal
}
