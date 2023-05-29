package server

import (
	"net/http"
)

// GET /health
func (s *Server) GetHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello world!"))
}
