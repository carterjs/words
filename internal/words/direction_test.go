package words_test

import (
	"github.com/carterjs/words/internal/words"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirection_Other(t *testing.T) {
	assert.Equal(t, words.DirectionHorizontal.Other(), words.DirectionVertical)
	assert.Equal(t, words.DirectionVertical.Other(), words.DirectionHorizontal)
}
