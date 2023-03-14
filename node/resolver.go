package node

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/shekarramaswamy4/shared-register-abstraction/shared"
)

func (n *Node) StartHTTP() {
	log.Printf("Running node %d on port %d\n", n.ID, n.Port)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", n.Port),
		Handler: n,
	}
	n.Server = server

	if err := server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			log.Printf("Server shut down: %s\n", err.Error())
		}
	}
}

func (n *Node) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Node %d received request: %s\n", n.ID, r.URL.Path)

	switch r.URL.Path {
	case "/read":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		vv, shouldInclude, err := n.ReadResolver(w, r)
		if err != nil {
			shared.WriteError(w, err)
			return
		}

		res := shared.NodeReadRes{
			ValueVersion:  vv,
			ShouldInclude: shouldInclude,
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			shared.WriteError(w, err)
		}

		return
	case "/write":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		shouldInclude, err := n.WriteResolver(w, r)
		if err != nil {
			shared.WriteError(w, err)
		}

		res := shared.NodeWriteRes{
			ShouldInclude: shouldInclude,
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
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

		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func (n *Node) ReadResolver(w http.ResponseWriter, r *http.Request) (shared.ValueVersion, bool, error) {
	addr := r.URL.Query().Get("address")
	return n.Read(addr)
}

func (n *Node) WriteResolver(w http.ResponseWriter, r *http.Request) (bool, error) {
	var req shared.WriteReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return false, err
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
