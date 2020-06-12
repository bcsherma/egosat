package egosat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndex(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(Lit(10).index(), 18)
	assert.Equal(Lit(-20).index(), 39)
}

func TestPolarity(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(Lit(10).polarity(), LTRUE)
	assert.Equal(Lit(-10).polarity(), LFALSE)
}
