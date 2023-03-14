package client

import (
	"testing"
	"time"

	"github.com/shekarramaswamy4/shared-register-abstraction/node"
	"github.com/stretchr/testify/assert"
)

func TestInitialization(t *testing.T) {
	n1 := node.New(8080)
	c := New(8070, 1, 8080)

	go n1.StartHTTP()
	go c.StartHTTP()
	time.Sleep(1)

	err := c.Write("addr1", "val1")
	assert.Nil(t, err)
	v, err := c.Read("addr1")
	assert.Nil(t, err)
	assert.Equal(t, "val1", v.Value)
	assert.Equal(t, 1, v.Version)

	n1.Server.Close()
	c.Server.Close()
}
