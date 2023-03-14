package node

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shekarramaswamy4/shared-register-abstraction/shared"
)

const pendingTimeout = 2 * time.Second

type Node struct {
	ID      string
	Port    int
	Memory  map[string]AddressData
	mutexes sync.Map
}

// AddressData is what's stored at each address
type AddressData struct {
	ValueVersion shared.ValueVersion

	PendingValue     *string
	PendingTimestamp *time.Time
}

func New(port int) *Node {
	return &Node{
		ID:     uuid.NewString(),
		Port:   port,
		Memory: make(map[string]AddressData),
	}
}

// Read returns the value at the given address
func (n *Node) Read(addr string) (shared.ValueVersion, error) {
	log.Printf("Node %s reading address %s", n.ID, addr)

	ad, ok := n.Memory[addr]
	if !ok || ad.ValueVersion.Version == 0 {
		// TODO: fractions - read from other nodes. Start typing the error
		return shared.ValueVersion{}, errors.New(fmt.Sprintf("Address %s not found", addr))
	}
	return ad.ValueVersion, nil
}

// Write "pre-commits" the specified value at the given address
func (n *Node) Write(addr string, val string) error {
	log.Printf("Node %s writing to address %s with value %s", n.ID, addr, val)

	loadMtx, _ := n.mutexes.LoadOrStore(addr, &sync.Mutex{})
	mtx := loadMtx.(*sync.Mutex)
	mtx.Lock()

	defer mtx.Unlock()

	ad, ok := n.Memory[addr]

	now := time.Now().UTC()
	// Current address has never been seen before
	if !ok {
		// TODO: fractions - determine if this node should have the address at all
		n.Memory[addr] = AddressData{
			ValueVersion:     shared.ValueVersion{},
			PendingValue:     &val,
			PendingTimestamp: &now,
		}
		log.Printf("Node %s precommited to address %s with value %s", n.ID, addr, val)
		return nil
	}

	// Current address has been seen before
	// No pending values for the current address
	if ad.PendingValue == nil {
		n.Memory[addr] = AddressData{
			ValueVersion:     ad.ValueVersion,
			PendingValue:     &val,
			PendingTimestamp: &now,
		}
		log.Printf("Node %s precommited to address %s with value %s", n.ID, addr, val)
	} else {
		// There is already a pending value for the current address
		pt := *ad.PendingTimestamp
		pv := *ad.PendingValue
		// timeout expired, replace!
		if pt.Add(pendingTimeout).Before(now) {
			n.Memory[addr] = AddressData{
				ValueVersion:     ad.ValueVersion,
				PendingValue:     &val,
				PendingTimestamp: &now,
			}

			log.Printf("Node %s precommited to address %s with value %s. Invalidated prev value %v", n.ID, addr, val, pv)
		} else {
			log.Printf("Node %s rejected precommitment to address %s with value %s. Pending value %v", n.ID, addr, val, pv)

			// timeout didn't expire, reject
			return errors.New(fmt.Sprintf("Address %s has a pending value %s", addr, *ad.PendingValue))
		}
	}

	return nil
}

// Confirm confirms the pending value at the given address
func (n *Node) Confirm(addr string) error {
	log.Printf("Node %s confirming address %s", n.ID, addr)

	loadMtx, _ := n.mutexes.LoadOrStore(addr, &sync.Mutex{})
	mtx := loadMtx.(*sync.Mutex)
	mtx.Lock()

	defer mtx.Unlock()

	ad, ok := n.Memory[addr]
	if !ok {
		return errors.New(fmt.Sprintf("Address %s not found", addr))
	}

	if ad.PendingValue == nil {
		return errors.New(fmt.Sprintf("Address %s has no pending value", addr))
	}

	version := ad.ValueVersion.Version + 1
	n.Memory[addr] = AddressData{
		ValueVersion: shared.ValueVersion{
			Value:   *ad.PendingValue,
			Version: version,
		},
		PendingValue:     nil,
		PendingTimestamp: nil,
	}

	log.Printf("Node %s confirmed address %s with value %s and version %d", n.ID, addr, *ad.PendingValue, version)

	return nil
}

// Update forcibly updates the current value and version at an address.
func (n *Node) Update(addr, val string, version int) error {
	log.Printf("Node %s updating address %s with val %s and version %d", n.ID, addr, val, version)

	loadMtx, _ := n.mutexes.LoadOrStore(addr, &sync.Mutex{})
	mtx := loadMtx.(*sync.Mutex)
	mtx.Lock()

	defer mtx.Unlock()

	ad, ok := n.Memory[addr]
	if !ok {
		return errors.New(fmt.Sprintf("Address %s not found", addr))
	}

	n.Memory[addr] = AddressData{
		ValueVersion: shared.ValueVersion{
			Value:   val,
			Version: version,
		},
		PendingValue:     ad.PendingValue,
		PendingTimestamp: ad.PendingTimestamp,
	}

	log.Printf("Node %s updated address %s with val %s and version %d", n.ID, addr, val, version)

	return nil
}
