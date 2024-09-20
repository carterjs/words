package words

type PlacementResult struct {
	LettersUsed   map[Point]rune
	DirectWord    Word
	IndirectWords []Word
	Modifiers     map[int]Modifier
	Points        int
}

func (placementResult PlacementResult) withComputedPoints(letterPoints map[rune]int) PlacementResult {
	score := scoreWord(placementResult.DirectWord, letterPoints, placementResult.Modifiers)
	for _, w := range placementResult.IndirectWords {
		wordScore := scoreWord(w, letterPoints, nil)
		score += wordScore
	}

	placementResult.Points = score

	return placementResult
}

func scoreWord(w Word, letterPoints map[rune]int, modifiers map[int]Modifier) int {
	var score int

	for i, letter := range w.Letters {
		point, _, _ := w.Index(i)
		if _, isBlank := w.Blanks[point]; isBlank {
			continue
		}

		letterScore := letterPoints[letter]

		if modifier, hasModifier := modifiers[i]; hasModifier {
			letterScore = modifier.ModifyLetterScore(letterScore)
		}

		score += letterScore
	}

	var wordModifiers []string
	for _, modifier := range modifiers {
		before := score
		score = modifier.ModifyWordScore(score)
		if before != score {
			wordModifiers = append(wordModifiers, string(modifier))
		}
	}

	return score
}
