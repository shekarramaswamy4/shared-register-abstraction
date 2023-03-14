package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIncludeInShard(t *testing.T) {
	assert.True(t, includeInShard(0, 1, 3, 2))
	assert.True(t, includeInShard(3, 0, 3, 2))
	assert.True(t, includeInShard(3, 1, 3, 2))
	assert.True(t, includeInShard(5, 1, 3, 3))
	assert.True(t, includeInShard(4, 0, 5, 2))

	assert.False(t, includeInShard(0, 2, 3, 2))
	assert.False(t, includeInShard(2, 1, 3, 2))
	assert.False(t, includeInShard(4, 1, 5, 2))
}
