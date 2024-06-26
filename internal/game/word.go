package game

import "fmt"

type Word struct {
	X         int
	Y         int
	Direction Direction
	Value     string
}

type Direction string

const (
	DirectionHorizontal Direction = "horizontal"
	DirectionVertical   Direction = "vertical"
)

func (direction Direction) Other() Direction {
	if direction == DirectionHorizontal {
		return DirectionVertical
	}

	return DirectionHorizontal

}

func (word Word) String() string {
	return fmt.Sprintf("%q@%d,%d", word.Value, word.X, word.Y)
}
