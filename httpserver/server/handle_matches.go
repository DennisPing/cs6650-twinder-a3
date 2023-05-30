package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

// GET /matches/{userId}/
func (s *Server) GetMatches(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, fmt.Sprintf("invalid userId: %s", userId))
		return
	}
	matches, err := s.store.GetMatches(r.Context(), userIdInt)
	if err != nil {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, err.Error())
		return
	}
	writeJsonResponse(w, r.Method, http.StatusOK, matches)
}
