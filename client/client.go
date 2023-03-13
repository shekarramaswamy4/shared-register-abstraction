package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shekarramaswamy4/shared-register-abstraction/shared"
)

type Client struct {
	ID              string
	NumNodes        int
	QuorumThreshold int
	NodePorts       []string
	httpClient      http.Client
}

func New(numNodes int, firstNodePort int) *Client {
	c := &Client{
		ID:              uuid.NewString(),
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
	ValueVersion shared.ValueVersion
	Port         string
	Err          error
}

func (c *Client) Read(addr string) (shared.ValueVersion, error) {
	ch := make(chan readResult)

	// Read from the nodes in parallel
	for _, port := range c.NodePorts {
		port := port
		go func(port string) {
			vv, err := c.readFromNode(addr, port)
			ch <- readResult{vv, port, err}
		}(port)
	}

	// Collect the results
	var readRes []readResult
	for i := 0; i < len(c.NodePorts); i++ {
		res := <-ch
		if res.Err != nil {
			fmt.Printf("Error reading from node %s: %s", res.Port, res.Err)
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
		}

		validResponses++
		if currentValue == nil {
			currentValue = &res.ValueVersion.Value
			latestVersion = &res.ValueVersion.Version
		} else if res.ValueVersion.Version > *latestVersion {
			currentValue = &res.ValueVersion.Value
			latestVersion = &res.ValueVersion.Version
		}
	}

	if validResponses < c.QuorumThreshold {
		return shared.ValueVersion{}, fmt.Errorf("Fetching from quorum not reached")
	}

	// TODO: update stale nodes?

	return shared.ValueVersion{
		Value:   *currentValue,
		Version: *latestVersion,
	}, nil
}

func (c *Client) Write(addr string, val string) error {
	return nil
}

func (c *Client) readFromNode(addr string, port string) (shared.ValueVersion, error) {
	resp, err := c.httpClient.Get(shared.CreateURL(port, "/read?address="+addr))
	if err != nil {
		return shared.ValueVersion{}, err
	}

	var vv shared.ValueVersion
	if err := json.NewDecoder(resp.Body).Decode(&vv); err != nil {
		return shared.ValueVersion{}, err
	}

	return vv, nil
}
