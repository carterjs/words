package words

import (
	"fmt"
	"strings"
)

type PlacementResult struct {
	LettersUsed       []rune
	DirectWord        Word
	IndirectWords     []Word
	Modifiers         map[int]Modifier
	Points            int
	PointsExplanation string
}

func (placementResult PlacementResult) withComputedPoints(letterPoints map[rune]int) PlacementResult {
	score, explanation := scoreWord(placementResult.DirectWord, letterPoints, placementResult.Modifiers)
	for _, w := range placementResult.IndirectWords {
		wordScore, wordExplanation := scoreWord(w, letterPoints, nil)
		score += wordScore
		explanation += " + " + wordExplanation
	}

	placementResult.Points = score
	placementResult.PointsExplanation = explanation

	return placementResult
}

func scoreWord(w Word, letterPoints map[rune]int, modifiers map[int]Modifier) (int, string) {
	var score int
	explanation := fmt.Sprintf("%s=", w.String())

	for i, letter := range w.Letters {
		if i > 0 {
			explanation += "+"
		}

		if _, isBlank := w.Blanks[i]; isBlank {
			explanation += fmt.Sprintf("%d", letterPoints[BlankLetter])
			continue
		}

		letterScore := letterPoints[letter]
		if modifier, hasModifier := modifiers[i]; hasModifier {
			before := letterScore
			letterScore = modifier.ModifyLetterScore(letterScore)
			if before != letterScore {
				explanation += fmt.Sprintf("%s%d=", string(modifier), before)
			}
		}

		explanation += fmt.Sprint(letterScore)
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

	if len(wordModifiers) > 0 {
		explanation = fmt.Sprintf("(%s)(%s)", strings.Join(wordModifiers, "*"), explanation)
	}

	explanation += fmt.Sprintf(" = %d", score)
	return score, explanation
}
