package node

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/shekarramaswamy4/shared-register-abstraction/shared"
)

const pendingTimeout = 2 * time.Second

type Node struct {
	Server     *http.Server
	ID         int
	Port       int
	TotalNodes int
	// NumReplicas / TotalNodes defines the fraction of nodes values should be replicated to
	// (TotalNodes/2+1) <= NumReplicas <= TotalNodes
	NumReplicas int

	Memory  map[string]AddressData
	mutexes sync.Map

	Flags TestingFlags
}

type TestingFlags struct {
	RefuseRead    bool
	RefuseWrite   bool
	RefuseConfirm bool
	Time          *time.Time
}

// AddressData is what's stored at each address
type AddressData struct {
	ValueVersion shared.ValueVersion

	PendingValue     *string
	PendingTimestamp *time.Time
}

func New(id, port, totalNodes, numReplicas int) *Node {
	return &Node{
		ID:          id,
		Port:        port,
		TotalNodes:  totalNodes,
		NumReplicas: numReplicas,

		Memory: make(map[string]AddressData),

		Flags: TestingFlags{},
	}
}

func (n *Node) GetNow() time.Time {
	if n.Flags.Time != nil {
		return *n.Flags.Time
	}
	return time.Now().UTC()
}

// Read returns the value at the given address
func (n *Node) Read(addr string) (shared.ValueVersion, bool, error) {
	log.Printf("Node %d reading address %s", n.ID, addr)

	if n.Flags.RefuseRead {
		return shared.ValueVersion{}, false, errors.New("Refusing to read because of testing flag")
	}

	shouldInclude := shared.HashAndCheckShardInclusion(addr, n.ID, n.TotalNodes, n.NumReplicas)
	if !shouldInclude {
		return shared.ValueVersion{}, false, nil
	}

	ad, ok := n.Memory[addr]
	if !ok || ad.ValueVersion.Version == 0 {
		return shared.ValueVersion{}, true, errors.New(fmt.Sprintf("Address %s not found", addr))
	}

	log.Printf("Node %d returned address %s with value %s and version %d", n.ID, addr, ad.ValueVersion.Value, ad.ValueVersion.Version)
	return ad.ValueVersion, true, nil
}

// Write "pre-commits" the specified value at the given address
func (n *Node) Write(addr string, val string) (bool, error) {
	log.Printf("Node %d writing to address %s with value %s", n.ID, addr, val)

	if n.Flags.RefuseWrite {
		return false, errors.New("Refusing to write because of testing flag")
	}

	shouldInclude := shared.HashAndCheckShardInclusion(addr, n.ID, n.TotalNodes, n.NumReplicas)
	if !shouldInclude {
		return false, nil
	}

	loadMtx, _ := n.mutexes.LoadOrStore(addr, &sync.Mutex{})
	mtx := loadMtx.(*sync.Mutex)
	mtx.Lock()

	defer mtx.Unlock()

	ad, ok := n.Memory[addr]

	now := n.GetNow()
	// Current address has never been seen before
	if !ok {
		n.Memory[addr] = AddressData{
			ValueVersion:     shared.ValueVersion{},
			PendingValue:     &val,
			PendingTimestamp: &now,
		}
		log.Printf("Node %d precommited to address %s with value %s", n.ID, addr, val)
		return true, nil
	}

	// Current address has been seen before
	// No pending values for the current address
	if ad.PendingValue == nil {
		n.Memory[addr] = AddressData{
			ValueVersion:     ad.ValueVersion,
			PendingValue:     &val,
			PendingTimestamp: &now,
		}
		log.Printf("Node %d precommited to address %s with value %s", n.ID, addr, val)
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

			log.Printf("Node %d precommited to address %s with value %s. Invalidated prev value %v", n.ID, addr, val, pv)
		} else {
			log.Printf("Node %d rejected precommitment to address %s with value %s at time %v. Pending value %v at time %v", n.ID, addr, val, now, pv, pt)

			// timeout didn't expire, reject
			return true, errors.New(fmt.Sprintf("Address %s has a pending value %s", addr, *ad.PendingValue))
		}
	}

	return true, nil
}

// Confirm confirms the pending value at the given address
func (n *Node) Confirm(addr string) error {
	log.Printf("Node %d confirming address %s", n.ID, addr)

	if n.Flags.RefuseConfirm {
		return errors.New("Refusing to confirm because of testing flag")
	}

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

	log.Printf("Node %d confirmed address %s with value %s and version %d", n.ID, addr, *ad.PendingValue, version)

	return nil
}

// Update forcibly updates the current value and version at an address.
func (n *Node) Update(addr, val string, version int) error {
	log.Printf("Node %d updating address %s with val %s and version %d", n.ID, addr, val, version)

	loadMtx, _ := n.mutexes.LoadOrStore(addr, &sync.Mutex{})
	mtx := loadMtx.(*sync.Mutex)
	mtx.Lock()

	defer mtx.Unlock()

	updatedVV := shared.ValueVersion{
		Value:   val,
		Version: version,
	}

	ad, ok := n.Memory[addr]
	if !ok {
		n.Memory[addr] = AddressData{
			ValueVersion: updatedVV,
		}
		return nil
	}

	n.Memory[addr] = AddressData{
		ValueVersion:     updatedVV,
		PendingValue:     ad.PendingValue,
		PendingTimestamp: ad.PendingTimestamp,
	}

	log.Printf("Node %d updated address %s with val %s and version %d", n.ID, addr, val, version)

	return nil
}
