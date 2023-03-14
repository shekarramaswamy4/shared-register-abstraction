package client

import (
	"testing"

	"github.com/shekarramaswamy4/shared-register-abstraction/node"
	"github.com/stretchr/testify/assert"
)

func TestInitialization(t *testing.T) {
	n1 := node.New(8080)
	n1.StartHTTP()
	c := New(8070, 1, 8080)
	c.StartHTTP()

	_, err := c.Read("addr1")
	assert.Nil(t, err)

	err = c.Write("addr1", "val1")
	assert.Nil(t, err)
	v, err := c.Read("addr1")
	assert.Nil(t, err)
	assert.Equal(t, v.Value, "val1")
	assert.Equal(t, v.Version, 1)
}
