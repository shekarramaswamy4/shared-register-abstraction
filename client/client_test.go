package client

import (
	"testing"
	"time"

	"github.com/shekarramaswamy4/shared-register-abstraction/node"
	"github.com/stretchr/testify/assert"
)

func TestInitialization(t *testing.T) {
	n1 := node.New(0, 8080, 1, 1)
	c := New(8070, 1, 8080)

	go n1.StartHTTP()
	go c.StartHTTP()

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
	n1 := node.New(0, 8080, 1, 1)
	n2 := node.New(1, 8081, 1, 1)
	n3 := node.New(2, 8082, 1, 1)

	c := New(8070, 3, 8080)

	go n1.StartHTTP()
	go n2.StartHTTP()
	go n3.StartHTTP()
	go c.StartHTTP()

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

// TestOneDownNodeAndForceUpdate tests the case where one node out of three is down.
// Writes should still confirm, and a read should update the node where the write failed.
func TestOneDownNodeAndForceUpdate(t *testing.T) {
	n1 := node.New(0, 8080, 1, 1)
	n2 := node.New(1, 8081, 1, 1)
	n3 := node.New(2, 8082, 1, 1)

	n1.Flags.RefuseWrite = true

	c := New(8070, 3, 8080)

	go n1.StartHTTP()
	go n2.StartHTTP()
	go n3.StartHTTP()
	go c.StartHTTP()

	err := c.Write("addr1", "val1")
	assert.Nil(t, err)

	v, err := n1.Read("addr1")
	// n1 doesn't have anything written to it yet
	assert.NotNil(t, err)
	v, err = n2.Read("addr1")
	assert.Nil(t, err)
	assert.Equal(t, "val1", v.Value)
	assert.Equal(t, 1, v.Version)
	v, err = n3.Read("addr1")
	assert.Nil(t, err)
	assert.Equal(t, "val1", v.Value)
	assert.Equal(t, 1, v.Version)

	v, err = c.Read("addr1")
	assert.Nil(t, err)
	assert.Equal(t, "val1", v.Value)
	assert.Equal(t, 1, v.Version)

	// After reading, the client should update other members to get up to speed
	// Therefore, n1 should be updated
	v, err = n1.Read("addr1")
	assert.Nil(t, err)
	assert.Equal(t, "val1", v.Value)
	assert.Equal(t, 1, v.Version)

	n1.Server.Close()
	n2.Server.Close()
	n3.Server.Close()
	c.Server.Close()
}

func TestNoQuorumWrites(t *testing.T) {
	n1 := node.New(0, 8080, 1, 1)
	n2 := node.New(1, 8081, 1, 1)
	n3 := node.New(2, 8082, 1, 1)

	n1.Flags.RefuseWrite = true
	n2.Flags.RefuseWrite = true

	c := New(8070, 3, 8080)

	go n1.StartHTTP()
	go n2.StartHTTP()
	go n3.StartHTTP()
	go c.StartHTTP()

	err := c.Write("addr1", "val1")
	assert.NotNil(t, err)

	n1.Server.Close()
	n2.Server.Close()
	n3.Server.Close()
	c.Server.Close()
}

// Test that values aren't written if there's no confirmation.
// Test a subsequent write to the same address before and after the threshold
func TestNoQuorumConfirms(t *testing.T) {
	n1 := node.New(0, 8080, 1, 1)
	n2 := node.New(1, 8081, 1, 1)
	n3 := node.New(2, 8082, 1, 1)

	n1.Flags.RefuseConfirm = true
	n2.Flags.RefuseConfirm = true

	c := New(8070, 3, 8080)

	go n1.StartHTTP()
	go n2.StartHTTP()
	go n3.StartHTTP()
	go c.StartHTTP()

	// Should fail b/c of no confirmations
	err := c.Write("addr1", "val1")
	assert.NotNil(t, err)

	n1.Flags.RefuseConfirm = false
	n2.Flags.RefuseConfirm = false
	// Should fail b/c of the pending confirmations
	err = c.Write("addr1", "val2")
	assert.NotNil(t, err)
	// Should succeed b/c different keys
	err = c.Write("addr2", "val3")
	assert.Nil(t, err)
	v, err := c.Read("addr2")
	assert.Nil(t, err)
	assert.Equal(t, "val3", v.Value)
	assert.Equal(t, 1, v.Version)

	forwardTime := time.Now().UTC().Add(time.Second * 3)
	n1.Flags.Time = &forwardTime
	n2.Flags.Time = &forwardTime
	// Should succeed since the pending confirmations have expired
	err = c.Write("addr1", "val2")
	assert.Nil(t, err)

	n1.Server.Close()
	n2.Server.Close()
	n3.Server.Close()
	c.Server.Close()
}

// Test one client can read another's writes
// Test the other client can update the version and the other client can read it
func TestMultiClientBasic(t *testing.T) {
	n1 := node.New(0, 8080, 1, 1)
	n2 := node.New(1, 8081, 1, 1)
	n3 := node.New(2, 8082, 1, 1)

	c1 := New(8070, 3, 8080)
	c2 := New(8070, 3, 8080)

	go n1.StartHTTP()
	go n2.StartHTTP()
	go n3.StartHTTP()
	go c1.StartHTTP()
	go c2.StartHTTP()

	err := c1.Write("addr1", "val1")
	assert.Nil(t, err)
	v, err := c2.Read("addr1")
	assert.Nil(t, err)
	assert.Equal(t, "val1", v.Value)
	assert.Equal(t, 1, v.Version)

	err = c2.Write("addr1", "val3")
	assert.Nil(t, err)
	v, err = c1.Read("addr1")
	assert.Nil(t, err)
	assert.Equal(t, "val3", v.Value)
	assert.Equal(t, 2, v.Version)

	n1.Server.Close()
	n2.Server.Close()
	n3.Server.Close()
	c1.Server.Close()
	c2.Server.Close()
}
