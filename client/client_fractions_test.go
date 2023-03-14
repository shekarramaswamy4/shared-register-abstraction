package client

import (
	"testing"

	"github.com/shekarramaswamy4/shared-register-abstraction/node"
	"github.com/stretchr/testify/assert"
)

func TestReplicasBasic(t *testing.T) {
	// hash(addr1) = 1443559033, 1443559033 % 3 = 1
	// This means that addr1 should NOT be stored in node 0, but writes should go through.
	n1 := node.New(0, 8080, 3, 2)
	n2 := node.New(1, 8081, 3, 2)
	n3 := node.New(2, 8082, 3, 2)

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

	// Not stored on n1, so should return false
	_, shouldInclude, err := n1.Read("addr1")
	assert.False(t, shouldInclude)

	n1.Server.Close()
	n2.Server.Close()
	n3.Server.Close()
	c.Server.Close()
}

// Tests that a write cannot go through if 1/2 replicas in a cluster of 3 is down.
func TestReplicasWithFault(t *testing.T) {
	// hash(addr1) = 1443559033, 1443559033 % 3 = 1
	// This means that addr1 should NOT be stored in node 0, but writes should go through.
	n1 := node.New(0, 8080, 3, 2)
	n2 := node.New(1, 8081, 3, 2)
	n3 := node.New(2, 8082, 3, 2)

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

// Tests that if there's only one replica specified, then the write can't go through because
// quorum can't be achieved.
func TestInvalidReplicas(t *testing.T) {
	// hash(addr1) = 1443559033, 1443559033 % 3 = 1
	n1 := node.New(0, 8080, 3, 1)
	n2 := node.New(1, 8081, 3, 1)
	n3 := node.New(2, 8082, 3, 1)

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

// Tests that in a cluster of 5 and 4 replicas, if one replica is down the write still goes through
func TestQuorumWithOneFaultyReplica(t *testing.T) {
	// hash(addr1) = 1443559033, 1443559033 % 3 = 3
	// This means that addr1 should NOT be stored in node 2, but writes should go through.
	n1 := node.New(0, 8080, 5, 4)
	n2 := node.New(1, 8081, 5, 4)
	n3 := node.New(2, 8082, 5, 4)
	n4 := node.New(3, 8083, 5, 4)
	n5 := node.New(4, 8084, 5, 4)

	n4.Flags.RefuseWrite = true

	c := New(8070, 5, 8080)

	go n1.StartHTTP()
	go n2.StartHTTP()
	go n3.StartHTTP()
	go n4.StartHTTP()
	go n5.StartHTTP()
	go c.StartHTTP()

	// Write should still go through
	err := c.Write("addr1", "val1")
	assert.Nil(t, err)
	v, err := c.Read("addr1")
	assert.Nil(t, err)
	assert.Equal(t, "val1", v.Value)
	assert.Equal(t, 1, v.Version)

	// n3 should be updated
	vv, shouldInclude, err := n4.Read("addr1")
	assert.Nil(t, err)
	assert.True(t, shouldInclude)
	assert.Equal(t, "val1", vv.Value)

	n1.Server.Close()
	n2.Server.Close()
	n3.Server.Close()
	n4.Server.Close()
	n5.Server.Close()
	c.Server.Close()
}
