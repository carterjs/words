package words

import "github.com/carterjs/words/internal/pattern"

var StandardConfig = Config{
	LetterDistribution: map[rune]int{
		'A':         9,
		'B':         2,
		'C':         2,
		'D':         4,
		'E':         12,
		'F':         2,
		'G':         3,
		'H':         2,
		'I':         9,
		'J':         1,
		'K':         1,
		'L':         4,
		'M':         2,
		'N':         6,
		'O':         8,
		'P':         2,
		'Q':         1,
		'R':         6,
		'S':         4,
		'T':         6,
		'U':         4,
		'V':         2,
		'W':         2,
		'X':         1,
		'Y':         2,
		'Z':         1,
		BlankLetter: 2,
	},
	LetterPoints: map[rune]int{
		'A':         1,
		'B':         3,
		'C':         3,
		'D':         2,
		'E':         1,
		'F':         4,
		'G':         2,
		'H':         4,
		'I':         1,
		'J':         8,
		'K':         5,
		'L':         1,
		'M':         3,
		'N':         1,
		'O':         1,
		'P':         3,
		'Q':         10,
		'R':         1,
		'S':         1,
		'T':         1,
		'U':         1,
		'V':         4,
		'W':         4,
		'X':         8,
		'Y':         4,
		'Z':         10,
		BlankLetter: 0,
	},
	Modifiers: pattern.Group[Modifier]{
		{
			Value: ModifierTripleWord,
			Grids: []pattern.Grid{
				{
					Width:  9,
					Height: 9,
				},
			},
		},
		{
			Value: ModifierDoubleWord,
			BothDiagonals: []pattern.BothDiagonals{
				{
					StartAt:    3,
					SkipCount:  2,
					MatchCount: 4,
				},
			},
		},
		{
			Value: ModifierTripleLetter,
			Grids: []pattern.Grid{
				{
					X:      2,
					Y:      2,
					Width:  5,
					Height: 5,
				},
				//{
				//	X:      -2,
				//	Y:      -3,
				//	Width:  7,
				//	Height: 7,
				//},
			},
		},
		{
			Value: ModifierDoubleLetter,
			BothDiagonals: []pattern.BothDiagonals{
				{
					StartAt:    0,
					SkipCount:  0,
					MatchCount: 3,
				},
			},
			Grids: []pattern.Grid{
				{
					Width:  5,
					Height: 5,
				},
				{
					X:      1,
					Y:      5,
					Width:  11,
					Height: 11,
				},
				{
					X:      -1,
					Y:      5,
					Width:  11,
					Height: 11,
				},
				{
					X:      5,
					Y:      1,
					Width:  11,
					Height: 11,
				},
				{
					X:      5,
					Y:      -1,
					Width:  11,
					Height: 11,
				},
			},
		},
	},
	RackSize: 7,
}
