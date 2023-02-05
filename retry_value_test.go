package gokogeri

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsJSON(t *testing.T) {
	assert := require.New(t)

	assert.True(isJSONFalse(jsonFalse))
	assert.True(isJSONTrue(jsonTrue))

	assert.False(isJSONFalse(jsonTrue))
	assert.False(isJSONTrue(jsonFalse))
}
