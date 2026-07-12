package words

// Modifier multiplies letter or word scores at particular board cells.
type Modifier string

const (
	// ModifierDoubleLetter doubles the score of the letter placed on it.
	ModifierDoubleLetter = Modifier("DL")
	// ModifierTripleLetter triples the score of the letter placed on it.
	ModifierTripleLetter = Modifier("TL")
	// ModifierDoubleWord doubles the score of the whole word crossing it.
	ModifierDoubleWord = Modifier("DW")
	// ModifierTripleWord triples the score of the whole word crossing it.
	ModifierTripleWord = Modifier("TW")
)

const (
	doubleMultiplier = 2
	tripleMultiplier = 3
)

// ModifyLetterScore returns the letter score adjusted by the modifier.
func (modifier Modifier) ModifyLetterScore(score int) int {
	switch modifier {
	case ModifierDoubleLetter:
		return score * doubleMultiplier
	case ModifierTripleLetter:
		return score * tripleMultiplier
	default:
		return score
	}
}

// ModifyWordScore returns the word score adjusted by the modifier.
func (modifier Modifier) ModifyWordScore(score int) int {
	switch modifier {
	case ModifierDoubleWord:
		return score * doubleMultiplier
	case ModifierTripleWord:
		return score * tripleMultiplier
	default:
		return score
	}
}
