package words

import (
	_ "crypto/sha256"
	"github.com/carterjs/words/internal/pattern"
	"github.com/google/uuid"
	"math/rand"
	"sort"
)

type (
	Game struct {
		Started   bool
		ID        string
		Round     int
		Config    Config
		Pool      []rune
		PoolIndex int
		Players   []Player
		Turn      int
		Board     *Board
	}

	Config struct {
		LetterDistribution map[rune]int            `json:"letterDistribution"`
		LetterPoints       map[rune]int            `json:"letterPoints"`
		RackSize           int                     `json:"rackSize"`
		Modifiers          pattern.Group[Modifier] `json:"modifiers"`
	}
)

func NewGame(config Config, players ...string) *Game {
	id := uuid.NewString()

	game := &Game{
		ID:     id,
		Round:  1,
		Config: config,
		Pool:   config.getInitialLetterPool(),
		Board:  NewBoard(id, config),
	}

	for _, player := range players {
		game.Players = append(game.Players, *newPlayer(id, player))
	}

	return game
}

func (game *Game) AddPlayer(name string) (*Player, error) {
	if game.Started {
		return nil, ErrGameStarted
	}

	player := newPlayer(game.ID, name)
	game.Players = append(game.Players, *player)
	return player, nil
}

func (game *Game) Start() error {
	if len(game.Players) < 1 {
		return ErrNotEnoughPlayers
	}

	err := game.fillPlayerRacks()
	if err != nil {
		return err
	}

	game.Started = true

	return nil
}

func (game *Game) LettersRemaining() int {
	return len(game.Pool) - game.PoolIndex
}

func (game *Game) fillPlayerRacks() error {
	for i := range game.Players {
		err := game.fillPlayerRack(&game.Players[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (game *Game) fillPlayerRack(player *Player) error {
	needed := game.Config.RackSize - len(player.Letters)
	if needed > 0 {
		hand, err := game.takeLettersFromPool(needed)
		if err != nil {
			return err
		}

		player.giveLetters(hand)
	}

	return nil
}

func (config Config) getInitialLetterPool() []rune {
	var letters []rune
	for letter, count := range config.LetterDistribution {
		for i := 0; i < count; i++ {
			letters = append(letters, letter)
		}
	}

	// shuffle
	rand.Shuffle(len(letters), func(i, j int) {
		letters[i], letters[j] = letters[j], letters[i]
	})

	return letters
}

func (game *Game) takeLettersFromPool(n int) ([]rune, error) {
	if game.PoolIndex == len(game.Pool) {
		return nil, ErrNoLettersInPool
	}

	if n > len(game.Pool)-game.PoolIndex {
		n = len(game.Pool) - game.PoolIndex
	}

	letters := game.Pool[game.PoolIndex : game.PoolIndex+n]
	game.PoolIndex += n

	return letters, nil
}

func (game *Game) GetPlayerByID(id string) *Player {
	for _, player := range game.Players {
		if player.ID == id {
			return &player
		}
	}

	return nil
}

func (game *Game) GetPlayerByName(name string) *Player {
	for _, player := range game.Players {
		if player.Name == name {
			return &player
		}
	}

	return nil
}

func (game *Game) FindPlacements(playerID string, point Point, s string) ([]PlacementResult, error) {
	if !game.Started {
		return nil, ErrGameNotStarted
	}

	var placements []PlacementResult
	// check the word starting at each position for each direction
	for _, direction := range []Direction{DirectionHorizontal, DirectionVertical} {
		for i := range len(s) {
			word := NewWord(point.Offset(direction.Vector(-i)), direction, s)

			// check placement
			result, err := game.CheckWord(playerID, word)
			if err != nil {
				continue
			}

			placements = append(placements, result)
		}
	}

	if len(placements) == 0 {
		return nil, ErrCannotPlayWord
	}

	// stable sort by points
	sort.SliceStable(placements, func(i, j int) bool {
		return placements[i].Points > placements[j].Points
	})

	return placements, nil
}

func (game *Game) CheckWord(playerID string, word Word) (PlacementResult, error) {
	if !game.Started {
		return PlacementResult{}, ErrGameNotStarted
	}

	player := game.GetPlayerByID(playerID)
	if player == nil {
		return PlacementResult{}, ErrPlayerNotFound
	}

	result, err := game.Board.tryWordPlacement(word)
	if err != nil {
		return PlacementResult{}, err
	}

	canPlay, blanks := player.hasLettersWithBlanks(result.LettersUsed)
	if !canPlay {
		return PlacementResult{}, ErrCannotPlayWord
	}

	if len(blanks) > 0 {
		// recalculate placement result with blanks
		for point := range blanks {
			word = word.WithBlanks(point)
		}

		result, err = game.Board.tryWordPlacement(word)
		if err != nil {
			return PlacementResult{}, err
		}
	}

	// TODO: dictionary check

	return result, nil
}

func (game *Game) PlayWord(playerID string, word Word) (PlacementResult, error) {
	if !game.Started {
		return PlacementResult{}, ErrGameNotStarted
	}

	player := game.GetPlayerByID(playerID)
	if game.Players[game.Turn].ID != playerID {
		return PlacementResult{}, ErrNotYourTurn
	}

	result, err := game.CheckWord(playerID, word)
	if err != nil {
		return PlacementResult{}, err
	}

	player.takeLetters(lettersFromMap(result.LettersUsed))

	// place word on board
	result, err = game.Board.PlaceWord(word)
	if err != nil {
		return PlacementResult{}, err
	}

	err = game.fillPlayerRack(player)
	if err != nil {
		return PlacementResult{}, err
	}

	player.RecordResult(result)

	game.Players[game.Turn] = *player

	game.advanceTurn()

	return result, nil
}

func (game *Game) Undo() error {
	if !game.Started {
		return ErrGameNotStarted
	}

	lastTurn := game.Turn - 1
	if lastTurn < 0 {
		lastTurn = len(game.Players) - 1
		game.Round--
	}
	game.Turn = lastTurn

	err := game.Board.removeLastWord()
	if err != nil {
		return err
	}

	lastPlayer := &game.Players[lastTurn]

	// remove turn which fixes scores
	lastTurnResult := lastPlayer.Turns[len(lastPlayer.Turns)-1]
	lastPlayer.Turns = lastPlayer.Turns[:len(lastPlayer.Turns)-1]

	// put back same letters given to player
	lettersGiven := game.Pool[game.PoolIndex-len(lastTurnResult.LettersUsed) : game.PoolIndex]
	lastPlayer.takeLetters(lettersGiven)
	lastPlayer.giveLetters(lettersFromMap(lastTurnResult.LettersUsed))
	game.PoolIndex -= len(lastTurnResult.LettersUsed)

	// now shuffle everything after game.PoolIndex so that the player doesn't GetLetter the same letters again
	rand.Shuffle(len(game.Pool[game.PoolIndex:]), func(i, j int) {
		game.Pool[game.PoolIndex+i], game.Pool[game.PoolIndex+j] = game.Pool[game.PoolIndex+j], game.Pool[game.PoolIndex+i]
	})

	return nil
}

func lettersFromMap(m map[Point]rune) []rune {
	var letters []rune
	for _, letter := range m {
		letters = append(letters, letter)
	}
	return letters
}

func (game *Game) advanceTurn() {
	nextTurn := game.Turn + 1
	if nextTurn >= len(game.Players) {
		nextTurn = 0
		game.Round++
	}

	game.Turn = nextTurn
}
