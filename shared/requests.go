package shared

import "net/http"

func WriteError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

func CreateURL(port, path string) string {
	return "http://localhost:" + port + path
}

type WriteReq struct {
	Address string
	Value   string
}

type ConfirmReq struct {
	Address string
}

type UpdateReq struct {
	Address string
	Version int
	Value   string
}

type NodeReadRes struct {
	ValueVersion  ValueVersion
	ShouldInclude bool
}

type NodeWriteRes struct {
	ShouldInclude bool
}
