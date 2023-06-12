package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

// GET /stats/{userId}/
func (s *Server) GetStats(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, fmt.Sprintf("invalid userId: %s", userId))
		return
	}
	found, userStat, err := s.store.GetUserStats(context.Background(), userIdInt)
	if err != nil {
		writeErrorResponse(w, r.Method, http.StatusInternalServerError, err.Error())
		return
	}
	if !found {
		writeErrorResponse(w, r.Method, http.StatusNotFound, fmt.Sprintf("userId not found: %s", userId))
		return
	}
	writeJsonResponse(w, r.Method, http.StatusOK, userStat)
}
