package words

import "fmt"

// GameState is a serializable snapshot of a game, used by stores to persist
// and rebuild games. The board is not stored directly; it is rebuilt by
// replaying the words.
type GameState struct {
	ID             string               `json:"id"`
	Started        bool                 `json:"started"`
	Finished       bool                 `json:"finished"`
	Round          int                  `json:"round"`
	Turn           int                  `json:"turn"`
	ScorelessTurns int                  `json:"scorelessTurns"`
	Config         Config               `json:"config"`
	Pool           []rune               `json:"pool"`
	PoolIndex      int                  `json:"poolIndex"`
	Players        []PlayerState        `json:"players"`
	Words          []PlacedWordState    `json:"words"`
	LastWord       *LastPlacedWordState `json:"lastWord,omitempty"`
	Challenge      *ChallengeState      `json:"challenge,omitempty"`
	WinnerIDs      []string             `json:"winnerIds,omitempty"`
}

// PlayerState is a serializable snapshot of a player.
type PlayerState struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	Letters         []rune       `json:"letters"`
	Turns           []TurnRecord `json:"turns"`
	FinalAdjustment int          `json:"finalAdjustment"`
}

// PlacedWordState is a serializable snapshot of a placed word.
type PlacedWordState struct {
	Column    int       `json:"column"`
	Row       int       `json:"row"`
	Direction Direction `json:"direction"`
	Letters   string    `json:"letters"`
	Blanks    []Point   `json:"blanks,omitempty"`
}

// LastPlacedWordState is a serializable snapshot of the challenge window.
type LastPlacedWordState struct {
	PlayerID string `json:"playerId"`
	Settled  bool   `json:"settled"`
}

// ChallengeState is a serializable snapshot of an open challenge.
type ChallengeState struct {
	ChallengerID string          `json:"challengerId"`
	Votes        map[string]Vote `json:"votes"`
}

// State returns a snapshot of the game for persistence.
func (game *Game) State() GameState {
	state := GameState{
		ID:             game.id,
		Started:        game.started,
		Finished:       game.finished,
		Round:          game.round,
		Turn:           game.turn,
		ScorelessTurns: game.scorelessTurns,
		Config:         game.config,
		Pool:           game.pool,
		PoolIndex:      game.poolIndex,
		WinnerIDs:      game.winnerIDs,
	}

	for _, player := range game.players {
		state.Players = append(state.Players, PlayerState{
			ID:              player.id,
			Name:            player.name,
			Letters:         player.letters,
			Turns:           player.turns,
			FinalAdjustment: player.finalAdjustment,
		})
	}

	for _, word := range game.board.words {
		state.Words = append(state.Words, PlacedWordState{
			Column:    word.Start().Column(),
			Row:       word.Start().Row(),
			Direction: word.Direction(),
			Letters:   string(word.letters),
			Blanks:    word.Blanks(),
		})
	}

	if game.lastWord != nil {
		state.LastWord = &LastPlacedWordState{
			PlayerID: game.lastWord.playerID,
			Settled:  game.lastWord.settled,
		}
	}

	if game.challenge != nil {
		votes := make(map[string]Vote, len(game.challenge.votes))
		for playerID, vote := range game.challenge.votes {
			votes[playerID] = vote
		}

		state.Challenge = &ChallengeState{
			ChallengerID: game.challenge.challengerID,
			Votes:        votes,
		}
	}

	return state
}

// NewGameFromState rebuilds a game from a stored snapshot.
func NewGameFromState(state GameState) (*Game, error) {
	game := &Game{
		id:             state.ID,
		started:        state.Started,
		finished:       state.Finished,
		round:          state.Round,
		turn:           state.Turn,
		scorelessTurns: state.ScorelessTurns,
		config:         state.Config,
		pool:           state.Pool,
		poolIndex:      state.PoolIndex,
		winnerIDs:      state.WinnerIDs,
		board:          NewBoard(state.Config),
	}

	for _, playerState := range state.Players {
		game.players = append(game.players, Player{
			id:              playerState.ID,
			name:            playerState.Name,
			letters:         playerState.Letters,
			turns:           playerState.Turns,
			finalAdjustment: playerState.FinalAdjustment,
		})
	}

	for _, wordState := range state.Words {
		word := NewWord(NewPoint(wordState.Column, wordState.Row), wordState.Direction, wordState.Letters)
		word = word.WithBlanks(wordState.Blanks...)

		if _, err := game.board.PlaceWord(word); err != nil {
			return nil, fmt.Errorf("replaying stored word %q: %w", wordState.Letters, err)
		}
	}

	if state.LastWord != nil {
		game.lastWord = &lastWordRecord{
			playerID: state.LastWord.PlayerID,
			settled:  state.LastWord.Settled,
		}
	}

	if state.Challenge != nil {
		votes := make(map[string]Vote, len(state.Challenge.Votes))
		for playerID, vote := range state.Challenge.Votes {
			votes[playerID] = vote
		}

		game.challenge = &challengeRecord{
			challengerID: state.Challenge.ChallengerID,
			votes:        votes,
		}
	}

	return game, nil
}
