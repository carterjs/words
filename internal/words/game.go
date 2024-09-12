package words

import (
	"crypto"
	_ "crypto/sha256"
	"github.com/carterjs/words/internal/pattern"
	"github.com/google/uuid"
	"log/slog"
	"math/rand"
)

type (
	Game struct {
		ID             string
		PassphraseHash []byte
		Round          int
		Config         Config
		Pool           []rune
		PoolIndex      int
		Players        []Player
		Turn           int
		Board          *Board
	}

	Config struct {
		LetterDistribution map[rune]int
		LetterPoints       map[rune]int
		RackSize           int
		Modifiers          pattern.Group[Modifier]
	}
)

func NewGame(config Config, players ...string) (*Game, error) {
	if len(players) < 1 {
		return nil, ErrNotEnoughPlayers
	}

	id := uuid.NewString()

	game := &Game{
		ID:     id,
		Round:  1,
		Config: config,
		Pool:   config.getInitialLetterPool(),
		Board:  newBoard(id, config),
	}

	for _, player := range players {
		game.Players = append(game.Players, *newPlayer(id, player))
	}

	err := game.fillPlayerRacks()
	if err != nil {
		return nil, err
	}

	return game, nil
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

		slog.Debug("Giving letters to player", "player", player.Name, "newLetters", string(hand))
		player.giveLetters(hand)
	}

	return nil
}

func (game *Game) SetPassword(passphrase string) {
	game.PassphraseHash = hashPassphrase(passphrase)
}

func (game *Game) PassphraseMatches(input string) bool {
	inputHash := hashPassphrase(input)
	return string(inputHash) == string(game.PassphraseHash)
}

func hashPassphrase(passphrase string) []byte {
	hash := crypto.SHA256.New()
	hash.Write([]byte(passphrase))
	return hash.Sum(nil)
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

func (game *Game) CheckWord(playerID string, word Word) (PlacementResult, error) {
	player := game.GetPlayerByID(playerID)
	if player == nil {
		return PlacementResult{}, ErrPlayerNotFound
	}

	result, err := game.Board.tryWordPlacement(word)
	if err != nil {
		return PlacementResult{}, err
	}

	if !player.hasLetters(result.LettersUsed) {
		return PlacementResult{}, ErrCannotPlayWord
	}

	// TODO: dictionary check

	return result, nil
}

func (game *Game) PlayWord(word Word) (PlacementResult, error) {
	player := &game.Players[game.Turn]
	result, err := game.CheckWord(player.ID, word)
	if err != nil {
		return PlacementResult{}, err
	}

	player.takeLetters(result.LettersUsed)

	// place word on board
	result, err = game.Board.placeWord(word)
	if err != nil {
		return PlacementResult{}, err
	}

	err = game.fillPlayerRack(player)
	if err != nil {
		return PlacementResult{}, err
	}

	player.RecordResult(result)

	game.AdvanceTurn()

	return result, nil
}

func (game *Game) Undo() error {
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
	lastPlayer.giveLetters(lastTurnResult.LettersUsed)
	game.PoolIndex -= len(lastTurnResult.LettersUsed)

	// now shuffle everything after game.PoolIndex so that the player doesn't get the same letters again
	rand.Shuffle(len(game.Pool[game.PoolIndex:]), func(i, j int) {
		game.Pool[game.PoolIndex+i], game.Pool[game.PoolIndex+j] = game.Pool[game.PoolIndex+j], game.Pool[game.PoolIndex+i]
	})

	return nil
}

func (game *Game) AdvanceTurn() {
	nextTurn := game.Turn + 1
	if nextTurn >= len(game.Players) {
		nextTurn = 0
		game.Round++
	}

	game.Turn = nextTurn
}
