package game

import (
	"math/rand/v2"
)

type Configuration struct {
	ID          string
	OwnerID     string
	Title       string
	Description string
	Letters     []LetterConfiguration
}

type LetterConfiguration struct {
	ConfigurationID string
	Letter          rune
	Points          int
	Count           int
}

func (configuration Configuration) GetLetters() []rune {
	letters := make([]rune, 0)
	for _, letterConfig := range configuration.Letters {
		for i := 0; i < letterConfig.Count; i++ {
			letters = append(letters, letterConfig.Letter)
		}
	}

	rand.Shuffle(len(letters), func(i, j int) {
		letters[i], letters[j] = letters[j], letters[i]
	})

	return letters
}

func (configuration Configuration) GetScoreForWord(word Word) int {
	scoreMap := make(map[rune]int)
	for _, letterConfig := range configuration.Letters {
		scoreMap[letterConfig.Letter] = letterConfig.Points
	}

	score := 0
	for _, letter := range word.Letters {
		score += scoreMap[letter]
	}

	return score
}
