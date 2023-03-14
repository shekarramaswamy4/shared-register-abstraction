package node

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/shekarramaswamy4/shared-register-abstraction/shared"
)

func (n *Node) StartHTTP() {
	log.Printf("Running node %s on port %d\n", n.ID, n.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", n.Port), n)
}

func (n *Node) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Node %s received request: %s\n", n.ID, r.URL.Path)

	switch r.URL.Path {
	case "/read":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		vv, err := n.ReadResolver(w, r)
		if err != nil {
			shared.WriteError(w, err)
			return
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
	case "/update":
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if err := n.UpdateResolver(w, r); err != nil {
			shared.WriteError(w, err)
		}
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

func (n *Node) UpdateResolver(w http.ResponseWriter, r *http.Request) error {
	var req shared.UpdateReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return err
	}

	return n.Update(req.Address, req.Value, req.Version)
}
