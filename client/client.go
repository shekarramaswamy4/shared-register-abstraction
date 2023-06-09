package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shekarramaswamy4/shared-register-abstraction/shared"
)

type Client struct {
	Server          *http.Server
	ID              string
	Port            int
	NumNodes        int
	QuorumThreshold int
	NodePorts       []string
	httpClient      http.Client
}

func New(port int, numNodes int, firstNodePort int) *Client {
	c := &Client{
		ID:              uuid.NewString(),
		Port:            port,
		NumNodes:        numNodes,
		QuorumThreshold: numNodes/2 + 1,
		httpClient: http.Client{
			Timeout: 3 * time.Second,
		},
	}

	nodePorts := make([]string, numNodes)
	nodePorts[0] = fmt.Sprintf("%d", firstNodePort)
	for i := 1; i < numNodes; i++ {
		nodePorts[i] = fmt.Sprintf("%d", firstNodePort+i)
	}

	c.NodePorts = nodePorts
	return c
}

type readResult struct {
	ValueVersion      shared.ValueVersion
	NodeShouldInclude bool
	Port              string
	Err               error
}

func (c *Client) Read(addr string) (shared.ValueVersion, error) {
	ch := make(chan readResult)

	// Read from the nodes in parallel
	for _, port := range c.NodePorts {
		port := port
		go func(port string) {
			vv, shouldInclude, err := c.readFromNode(addr, port)
			ch <- readResult{ValueVersion: vv, NodeShouldInclude: shouldInclude, Port: port, Err: err}
		}(port)
	}

	// Collect the results
	var readRes []readResult
	for i := 0; i < len(c.NodePorts); i++ {
		res := <-ch
		if res.Err != nil {
			log.Printf("Error reading from node %s: %s", res.Port, res.Err)
		}
		// TODO: don't wait for all reads to complete
		readRes = append(readRes, res)
	}

	// Determining what version to return
	var currentValue *string
	var latestVersion *int
	validResponses := 0
	for _, res := range readRes {
		if res.Err != nil {
			continue
		} else if !res.NodeShouldInclude {
			log.Printf("Node on port %s doesn't accept read to address %s", res.Port, addr)
			continue
		}

		validResponses++
		value := res.ValueVersion.Value
		version := res.ValueVersion.Version
		if currentValue == nil || res.ValueVersion.Version > *latestVersion {
			currentValue = &value
			latestVersion = &version
		}
	}

	if validResponses < c.QuorumThreshold {
		return shared.ValueVersion{}, fmt.Errorf("Not enough valid responses to make quorum")
	}

	log.Printf("Client %s read address %s with value %s and version %d", c.ID, addr, *currentValue, *latestVersion)

	// Update nodes that were behind
	// Now that we know the latest version and value, we simply iterate through the read responses
	// again and update the nodes that either errored or had an out of date version
	wg := sync.WaitGroup{}
	for _, res := range readRes {
		res := res

		if res.Err != nil || res.ValueVersion.Version != *latestVersion {
			wg.Add(1)
			go func(port string) {
				defer wg.Done()
				if err := c.updateNode(addr, *currentValue, *latestVersion, port); err != nil {
					log.Printf("Error updating node %s: %s", port, err)
				}
			}(res.Port)
		}
	}

	wg.Wait()

	return shared.ValueVersion{
		Value:   *currentValue,
		Version: *latestVersion,
	}, nil
}

func (c *Client) Write(addr string, val string) error {
	if err := c.write(addr, val); err != nil {
		return err
	}

	// OPTIMIZATION: Only send confirmations to nodes that acked the write
	return c.confirm(addr)
}

type writeResult struct {
	NodeShouldInclude bool
	Err               error
}

func (c *Client) write(addr string, val string) error {
	log.Printf("Attempting to write value %s to address %s\n", val, addr)
	// First write, then confirm
	writeCh := make(chan writeResult)

	// Write to the nodes in parallel
	for _, port := range c.NodePorts {
		port := port
		go func(port string) {
			shouldInclude, err := c.writeToNode(addr, val, port)
			writeCh <- writeResult{NodeShouldInclude: shouldInclude, Err: err}
		}(port)
	}

	// Collect the results
	numSuccessWrites := 0
	for i := 0; i < len(c.NodePorts); i++ {
		// TODO: don't wait for all writes to complete
		res := <-writeCh
		if res.Err != nil {
			log.Printf("Error writing to node on port %s: %s", c.NodePorts[i], res.Err)
		} else if !res.NodeShouldInclude {
			log.Printf("Node on port %s doesn't accept write to address %s", c.NodePorts[i], addr)
		} else {
			numSuccessWrites++
		}
	}

	if numSuccessWrites < c.QuorumThreshold {
		return fmt.Errorf("Writing to quorum not reached, try again later")
	}

	log.Printf("Client %s reached quorum writing %s to address %s\n", c.ID, val, addr)

	return nil
}

func (c *Client) confirm(addr string) error {
	log.Printf("Attempting to confirm address %s\n", addr)

	confirmCh := make(chan error)

	// Write to the nodes in parallel
	for _, port := range c.NodePorts {
		port := port
		go func(port string) {
			err := c.confirmWithNode(addr, port)
			confirmCh <- err
		}(port)
	}

	// Collect the results
	numSuccessConfirms := 0
	for i := 0; i < len(c.NodePorts); i++ {
		// TODO: don't wait for all confirms to complete
		res := <-confirmCh
		if res != nil {
			log.Printf("Error writing to node: %s", res)
		} else {
			numSuccessConfirms++
		}
	}

	if numSuccessConfirms < c.QuorumThreshold {
		return fmt.Errorf("Confirming to quorum not reached, try again later")
	}

	log.Printf("Client %s reached quorum confirming to address %s\n", c.ID, addr)

	return nil
}

func (c *Client) readFromNode(addr string, port string) (shared.ValueVersion, bool, error) {
	resp, err := c.httpClient.Get(shared.CreateURL(port, "/read?address="+addr))
	if err != nil {
		return shared.ValueVersion{}, false, err
	}

	if resp.StatusCode != http.StatusOK {
		return shared.ValueVersion{}, false, fmt.Errorf("Read failed: %d", resp.StatusCode)
	}

	var res shared.NodeReadRes
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return shared.ValueVersion{}, false, err
	}

	return res.ValueVersion, res.ShouldInclude, nil
}

func (c *Client) writeToNode(addr string, val string, port string) (bool, error) {
	body, _ := json.Marshal(shared.WriteReq{
		Address: addr,
		Value:   val,
	})
	resp, err := c.httpClient.Post(shared.CreateURL(port, "/write"), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Write failed: %d", resp.StatusCode)
	}

	var res shared.NodeWriteRes
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return false, err
	}

	return res.ShouldInclude, nil
}

func (c *Client) confirmWithNode(addr string, port string) error {
	body, _ := json.Marshal(shared.ConfirmReq{
		Address: addr,
	})
	req, _ := http.NewRequest(http.MethodPut, shared.CreateURL(port, "/confirm"), bytes.NewBuffer(body))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Confirm failed: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) updateNode(addr string, val string, version int, port string) error {
	body, _ := json.Marshal(shared.UpdateReq{
		Address: addr,
		Value:   val,
		Version: version,
	})
	req, _ := http.NewRequest(http.MethodPut, shared.CreateURL(port, "/update"), bytes.NewBuffer(body))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Update node failed: %d", resp.StatusCode)
	}

	return nil
}
