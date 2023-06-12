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
	found, matches, err := s.store.GetMatches(r.Context(), userIdInt)
	if err != nil {
		writeErrorResponse(w, r.Method, http.StatusInternalServerError, err.Error())
		return
	}
	if !found {
		writeErrorResponse(w, r.Method, http.StatusNotFound, fmt.Sprintf("userId not found: %s", userId))
		return
	}
	writeJsonResponse(w, r.Method, http.StatusOK, matches)
}
