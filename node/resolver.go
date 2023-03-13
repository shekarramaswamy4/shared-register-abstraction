package node

import (
	"encoding/json"
	"net/http"

	"github.com/shekarramaswamy4/shared-register-abstraction/shared"
)

type NodeResolver struct {
	N *Node
}

type writeReq struct {
	Addr  string
	Value string
}

type confirmReq struct {
	Addr string
}

func (nr *NodeResolver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/read":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		vv, err := nr.Read(w, r)
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

		if err := nr.Write(w, r); err != nil {
			shared.WriteError(w, err)
		}

		return
	case "/confirm":
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if err := nr.Confirm(w, r); err != nil {
			shared.WriteError(w, err)
		}

		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func (nr *NodeResolver) Read(w http.ResponseWriter, r *http.Request) (shared.ValueVersion, error) {
	addr := r.URL.Query().Get("address")
	return nr.N.Read(addr)
}

func (nr *NodeResolver) Write(w http.ResponseWriter, r *http.Request) error {
	var req writeReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return err
	}

	return nr.N.Write(req.Addr, req.Value)
}

func (nr *NodeResolver) Confirm(w http.ResponseWriter, r *http.Request) error {
	var req confirmReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return err
	}

	return nr.N.Confirm(req.Addr)
}
