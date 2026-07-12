package words

// PlacementResult describes the outcome of placing a word: the letters the
// player must spend, the word as placed (including blank substitutions), any
// perpendicular words completed by the placement, the modifiers hit, and the
// total points scored.
type PlacementResult struct {
	LettersUsed   map[Point]rune
	DirectWord    Word
	IndirectWords []Word
	Modifiers     map[int]Modifier
	Points        int
}

func placementWithPoints(result PlacementResult, letterPoints map[rune]int) PlacementResult {
	score := scoreWord(result.DirectWord, letterPoints, result.Modifiers)
	for _, indirectWord := range result.IndirectWords {
		score += scoreWord(indirectWord, letterPoints, nil)
	}

	result.Points = score

	return result
}

func scoreWord(word Word, letterPoints map[rune]int, modifiers map[int]Modifier) int {
	var score int

	for position, letter := range word.letters {
		point, _, _ := word.Index(position)
		if word.Blank(point) {
			continue
		}

		letterScore := letterPoints[letter]

		if modifier, hasModifier := modifiers[position]; hasModifier {
			letterScore = modifier.ModifyLetterScore(letterScore)
		}

		score += letterScore
	}

	for _, modifier := range modifiers {
		score = modifier.ModifyWordScore(score)
	}

	return score
}
