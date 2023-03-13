package shared

import "net/http"

func WriteError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

func CreateURL(port, path string) string {
	return "http://localhost:" + port + path
}
