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

func Test3Nodes(t *testing.T) {
	n1 := node.New(8080)
	n2 := node.New(8081)
	n3 := node.New(8082)

	c := New(8070, 3, 8080)

	go n1.StartHTTP()
	go n2.StartHTTP()
	go n3.StartHTTP()
	go c.StartHTTP()
	time.Sleep(1)

	err := c.Write("addr1", "val1")
	assert.Nil(t, err)
	v, err := c.Read("addr1")
	assert.Nil(t, err)
	assert.Equal(t, "val1", v.Value)
	assert.Equal(t, 1, v.Version)

	n1.Server.Close()
	n2.Server.Close()
	n3.Server.Close()
	c.Server.Close()
}
