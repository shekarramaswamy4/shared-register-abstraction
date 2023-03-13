package client

import (
	"encoding/json"
	"net/http"

	"github.com/shekarramaswamy4/shared-register-abstraction/shared"
)

type ClientResolver struct {
	C *Client
}

type writeReq struct {
	Addr  string
	Value string
}

func (nr *ClientResolver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (cr *ClientResolver) Read(w http.ResponseWriter, r *http.Request) (shared.ValueVersion, error) {
	addr := r.URL.Query().Get("address")
	return cr.C.Read(addr)
}

func (nr *ClientResolver) Write(w http.ResponseWriter, r *http.Request) error {
	var req writeReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return err
	}

	return nr.C.Write(req.Addr, req.Value)
}
