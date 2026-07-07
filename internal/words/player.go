package words

import (
	"github.com/google/uuid"
)

// Player is a participant in a game, holding a rack of letters and a record
// of scored turns.
type Player struct {
	id              string
	name            string
	letters         []rune
	turns           []TurnRecord
	finalAdjustment int
}

// TurnRecord captures the scoring outcome of a single played word.
type TurnRecord struct {
	Points       int            `json:"points"`
	LettersUsed  map[Point]rune `json:"lettersUsed"`
	LettersDrawn int            `json:"lettersDrawn"`
}

func newPlayer(name string) Player {
	return Player{
		id:   uuid.NewString(),
		name: name,
	}
}

// ID returns the player's unique identifier.
func (player Player) ID() string {
	return player.id
}

// Name returns the player's display name.
func (player Player) Name() string {
	return player.name
}

// Letters returns the letters currently on the player's rack.
func (player Player) Letters() []rune {
	letters := make([]rune, len(player.letters))
	copy(letters, player.letters)
	return letters
}

// Turns returns the player's scored turns in play order.
func (player Player) Turns() []TurnRecord {
	turns := make([]TurnRecord, len(player.turns))
	copy(turns, player.turns)
	return turns
}

// Score returns the player's total score, including any end-of-game
// adjustment for letters left on the rack.
func (player Player) Score() int {
	score := player.finalAdjustment
	for _, turn := range player.turns {
		score += turn.Points
	}

	return score
}

func (player *Player) hasLettersWithBlanks(lettersUsed map[Point]rune) (bool, map[Point]rune) {
	remaining := letterCounts(player.letters)

	blanks := make(map[Point]rune)

	for point, letter := range lettersUsed {
		if remaining[letter] == 0 {
			if remaining[BlankLetter] == 0 {
				return false, nil
			}

			remaining[BlankLetter]--
			blanks[point] = BlankLetter
		} else {
			remaining[letter]--
		}
	}

	return true, blanks
}

func (player *Player) hasLetters(letters []rune) bool {
	remaining := letterCounts(player.letters)

	for _, letter := range letters {
		if remaining[letter] == 0 {
			return false
		}
		remaining[letter]--
	}

	return true
}

func (player *Player) giveLetters(letters []rune) {
	player.letters = append(player.letters, letters...)
}

func (player *Player) takeLetters(letters []rune) {
	remaining := letterCounts(player.letters)

	for _, letter := range letters {
		remaining[letter]--
	}

	var newLetters []rune
	for letter, count := range remaining {
		for range count {
			newLetters = append(newLetters, letter)
		}
	}

	player.letters = newLetters
}

func letterCounts(letters []rune) map[rune]int {
	counts := make(map[rune]int)
	for _, letter := range letters {
		counts[letter]++
	}

	return counts
}
