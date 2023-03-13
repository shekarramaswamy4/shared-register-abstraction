package node

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/shekarramaswamy4/shared-register-abstraction/shared"
)

func (n *Node) StartHTTP(port string) {
	fmt.Printf("Running node %s on port %s\n", n.ID, port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), n)
}

func (n *Node) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/read":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		vv, err := n.ReadResolver(w, r)
		if err != nil {
			shared.WriteError(w, err)
		}

		if err := json.NewEncoder(w).Encode(vv); err != nil {
			shared.WriteError(w, err)
		}

		return
	case "/write":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if err := n.WriteResolver(w, r); err != nil {
			shared.WriteError(w, err)
		}

		return
	case "/confirm":
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if err := n.ConfirmResolver(w, r); err != nil {
			shared.WriteError(w, err)
		}

		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func (n *Node) ReadResolver(w http.ResponseWriter, r *http.Request) (shared.ValueVersion, error) {
	addr := r.URL.Query().Get("address")
	return n.Read(addr)
}

func (n *Node) WriteResolver(w http.ResponseWriter, r *http.Request) error {
	var req shared.WriteReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return err
	}

	return n.Write(req.Address, req.Value)
}

func (n *Node) ConfirmResolver(w http.ResponseWriter, r *http.Request) error {
	var req shared.ConfirmReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return err
	}

	return n.Confirm(req.Address)
}
