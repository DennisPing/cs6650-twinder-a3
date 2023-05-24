package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

// GET /matches/{userId}/
func (s *Server) MatchesHandler(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	_, err := strconv.Atoi(userId)
	if err != nil {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, fmt.Sprintf("invalid userId: %s", userId))
		return
	}
}
