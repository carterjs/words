package words

type Modifier string

const (
	ModifierDoubleLetter = Modifier("DL")
	ModifierTripleLetter = Modifier("TL")
	ModifierDoubleWord   = Modifier("DW")
	ModifierTripleWord   = Modifier("TW")
)

func (modifier Modifier) ModifyLetterScore(score int) int {
	switch modifier {
	case ModifierDoubleLetter:
		return score * 2
	case ModifierTripleLetter:
		return score * 3
	default:
		return score
	}
}

func (modifier Modifier) ModifyWordScore(score int) int {
	switch modifier {
	case ModifierDoubleWord:
		return score * 2
	case ModifierTripleWord:
		return score * 3
	default:
		return score
	}
}
