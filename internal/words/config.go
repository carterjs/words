package words

import (
	"math/rand"

	"github.com/carterjs/words/internal/pattern"
)

// Config describes the rules a game is played with: the letters available,
// their point values, the rack size, and the board's modifier layout.
type Config struct {
	LetterDistribution map[rune]int            `json:"letterDistribution"`
	LetterPoints       map[rune]int            `json:"letterPoints"`
	RackSize           int                     `json:"rackSize"`
	Modifiers          pattern.Group[Modifier] `json:"modifiers"`
}

// ConfigOverrides carries per-game adjustments applied on top of a preset.
type ConfigOverrides struct {
	RackSize           int
	LetterDistribution map[rune]int
	LetterPoints       map[rune]int
}

func configWithOverrides(config Config, overrides ConfigOverrides) Config {
	config.LetterDistribution = overriddenCounts(config.LetterDistribution, overrides.LetterDistribution)
	config.LetterPoints = overriddenCounts(config.LetterPoints, overrides.LetterPoints)

	if overrides.RackSize > 0 {
		config.RackSize = overrides.RackSize
	}

	return config
}

// overriddenCounts copies base so shared preset maps are never mutated.
func overriddenCounts(base map[rune]int, overrides map[rune]int) map[rune]int {
	counts := make(map[rune]int, len(base))
	for letter, count := range base {
		counts[letter] = count
	}
	for letter, count := range overrides {
		counts[letter] = count
	}

	return counts
}

func initialLetterPool(config Config) []rune {
	var letters []rune
	for letter, count := range config.LetterDistribution {
		for range count {
			letters = append(letters, letter)
		}
	}

	rand.Shuffle(len(letters), func(first, second int) {
		letters[first], letters[second] = letters[second], letters[first]
	})

	return letters
}
