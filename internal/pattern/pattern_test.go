package pattern_test

import (
	"github.com/carterjs/words/internal/pattern"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPattern(t *testing.T) {
	t.Parallel()

	test := []struct {
		name     string
		pattern  pattern.Pattern[bool]
		expected []string
	}{
		{
			name: "spaced x",
			pattern: pattern.Pattern[bool]{
				Value: true,
				BothDiagonals: []pattern.BothDiagonals{
					{
						StartAt:    0,
						SkipCount:  1,
						MatchCount: 1,
					},
				},
			},
			expected: []string{
				"_________",
				"_X_____X_",
				"_________",
				"___X_X___",
				"_________",
				"___X_X___",
				"_________",
				"_X_____X_",
				"_________",
			},
		},
		{
			name: "pattern spaced x",
			pattern: pattern.Pattern[bool]{
				Value: true,
				BothDiagonals: []pattern.BothDiagonals{
					{
						StartAt:    0,
						SkipCount:  1,
						MatchCount: 2,
					},
				},
			},
			expected: []string{
				"X_______X",
				"_________",
				"__X___X__",
				"___X_X___",
				"_________",
				"___X_X___",
				"__X___X__",
				"_________",
				"X_______X",
			},
		},
		{
			name: "3x2 grid",
			pattern: pattern.Pattern[bool]{
				Value: true,
				Grids: []pattern.Grid{
					{
						Width:  2,
						Height: 1,
					},
				},
			},
			expected: []string{
				"X_X_X",
				"X_X_X",
				"X_X_X",
			},
		},
		{
			name: "2x3 grid",
			pattern: pattern.Pattern[bool]{
				Value: true,
				Grids: []pattern.Grid{
					{
						Width:  2,
						Height: 3,
					},
				},
			},
			expected: []string{
				"_________",
				"X_X_X_X_X",
				"_________",
				"_________",
				"X_X_X_X_X",
				"_________",
				"_________",
				"X_X_X_X_X",
				"_________",
			},
		},
		{
			name: "3x3 grid",
			pattern: pattern.Pattern[bool]{
				Value: true,
				Grids: []pattern.Grid{
					{
						Width:  3,
						Height: 3,
					},
				},
			},
			expected: []string{
				"_________",
				"_X__X__X_",
				"_________",
				"_________",
				"_X__X__X_",
				"_________",
				"_________",
				"_X__X__X_",
				"_________",
			},
		},
	}

	for _, test := range test {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assertPattern(t, test.pattern, test.expected)
		})
	}
}

func assertPattern(t *testing.T, p pattern.Pattern[bool], expected []string) {
	t.Helper()

	height := len(expected)
	width := len(expected[0])

	results := make([]string, 0, height)
	for y := -height / 2; y <= height/2; y++ {
		results = append(results, "")
		for x := -width / 2; x <= width/2; x++ {
			if b, _ := p.Get(x, y); b {
				results[y+height/2] += "X"
			} else {
				results[y+height/2] += "_"
			}
		}
	}

	assert.Equal(t, expected, results)
}