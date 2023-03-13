package node

import (
	"encoding/json"
	"net/http"
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
		res, err := nr.Read(w, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		w.Write([]byte(res))

		return
	case "/write":
		if err := nr.Write(w, r); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		return
	case "/confirm":
		if err := nr.Confirm(w, r); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func (nr *NodeResolver) Read(w http.ResponseWriter, r *http.Request) (string, error) {
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
