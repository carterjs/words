package words

type Direction string

const (
	DirectionHorizontal Direction = "HORIZONTAL"
	DirectionVertical   Direction = "VERTICAL"
)

func (direction Direction) Vector(magnitude int) (int, int) {
	if direction == DirectionHorizontal {
		return magnitude, 0
	} else {
		return 0, magnitude
	}
}

func (direction Direction) Other() Direction {
	if direction == DirectionHorizontal {
		return DirectionVertical
	}

	return DirectionHorizontal
}
