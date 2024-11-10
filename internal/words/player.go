package words

import (
	"github.com/google/uuid"
)

type (
	Player struct {
		ID      string
		GameID  string
		Name    string
		Letters []rune
		Turns   []PlacementResult
	}
)

func newPlayer(gameID string, name string) *Player {
	return &Player{
		ID:     uuid.NewString(),
		GameID: gameID,
		Name:   name,
	}
}

func (player *Player) hasLetters(letters []rune) bool {
	playerLetters := make(map[rune]int)
	for _, letter := range player.Letters {
		playerLetters[letter]++
	}

	for _, letter := range letters {
		if playerLetters[letter] == 0 {
			return false
		}
		playerLetters[letter]--
	}

	return true
}

func (player *Player) hasLettersWithBlanks(lettersUsed map[Point]rune) (bool, map[Point]rune) {
	playerLetters := make(map[rune]int)
	for _, letter := range player.Letters {
		playerLetters[letter]++
	}

	blanks := make(map[Point]rune)

	for point, letter := range lettersUsed {
		if playerLetters[letter] == 0 {
			if playerLetters[BlankLetter] == 0 {
				return false, nil
			}

			playerLetters[BlankLetter]--
			blanks[point] = BlankLetter
		} else {
			playerLetters[letter]--
		}
	}

	return true, blanks
}

func (player *Player) giveLetters(letters []rune) {
	player.Letters = append(player.Letters, letters...)
}

func (player *Player) takeLetters(letters []rune) {
	playerLetters := make(map[rune]int)
	for _, letter := range player.Letters {
		playerLetters[letter]++
	}

	for _, letter := range letters {
		playerLetters[letter]--
	}

	var newLetters []rune
	for letter, count := range playerLetters {
		for i := 0; i < count; i++ {
			newLetters = append(newLetters, letter)
		}
	}

	player.Letters = newLetters
}

func (player *Player) RecordResult(result PlacementResult) {
	player.Turns = append(player.Turns, result)
}

func (player *Player) Score() int {
	var score int
	for _, result := range player.Turns {
		score += result.Points
	}

	return score
}
