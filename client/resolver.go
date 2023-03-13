package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/shekarramaswamy4/shared-register-abstraction/shared"
)

func (c *Client) StartHTTP(port string) {
	fmt.Printf("Running client %s on port %s\n", c.ID, port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), c)
}

func (c *Client) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/write":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		return
	case "/read":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func (c *Client) ReadResolver(w http.ResponseWriter, r *http.Request) (shared.ValueVersion, error) {
	addr := r.URL.Query().Get("address")
	return c.Read(addr)
}

func (c *Client) WriteResolver(w http.ResponseWriter, r *http.Request) error {
	var req shared.WriteReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return err
	}

	return c.Write(req.Address, req.Value)
}
