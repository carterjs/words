package game

import "math"

type Word struct {
	X         int
	Y         int
	Direction Direction
	Letters   []rune
}

func (word Word) Get(i int) (int, int, rune) {
	dx, dy := word.Direction.Vector(i)
	x, y := word.X+dx, word.Y+dy

	if i < 0 || i >= len(word.Letters) {
		return x, y, 0
	}

	return x, y, rune(word.Letters[i])
}

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

func (word Word) String() string {
	return string(word.Letters)
}

type WordStats struct {
	Usages     int
	Approvals  int
	Rejections int
}

func (stats WordStats) Reputation() float64 {
	approvalRate := float64(stats.Approvals) / float64(stats.Usages)
	rejectionRate := float64(stats.Rejections) / float64(stats.Usages)

	unweightedReputation := approvalRate - rejectionRate
	usageWeighting := 1 - math.Exp(-float64(stats.Usages)/5)

	return unweightedReputation * usageWeighting
}
