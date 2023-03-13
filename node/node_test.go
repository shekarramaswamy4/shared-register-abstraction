package node

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitialization(t *testing.T) {
	n := New()

	_, err := n.Read("addr1")
	assert.NotNil(t, err)

	err = n.Write("addr1", "val1")
	assert.Nil(t, err)
	_, err = n.Read("addr1")
	assert.NotNil(t, err)
}

func TestWriteAndConfirm(t *testing.T) {
	n := New()

	err := n.Write("addr1", "val1")
	assert.Nil(t, err)

	err = n.Confirm("addr1")
	assert.Nil(t, err)

	vv, err := n.Read("addr1")
	assert.Nil(t, err)
	assert.Equal(t, vv.Value, "val1")
	assert.Equal(t, vv.Version, 1)
}

func TestWriteNoTimeout(t *testing.T) {
	n := New()

	err := n.Write("addr1", "val1")
	assert.Nil(t, err)

	err = n.Write("addr1", "val2")
	assert.NotNil(t, err)

	err = n.Confirm("addr1")
	assert.Nil(t, err)

	vv, err := n.Read("addr1")
	assert.Nil(t, err)
	assert.Equal(t, vv.Value, "val1")
	assert.Equal(t, vv.Version, 1)
}

func TestWriteWithTimeout(t *testing.T) {
	n := New()

	err := n.Write("addr1", "val1")
	assert.Nil(t, err)

	time.Sleep(pendingTimeout + 1*time.Second)

	err = n.Write("addr1", "val2")
	assert.Nil(t, err)

	err = n.Confirm("addr1")
	assert.Nil(t, err)

	vv, err := n.Read("addr1")
	assert.Nil(t, err)
	assert.Equal(t, vv.Value, "val2")
	assert.Equal(t, vv.Version, 1)
}
