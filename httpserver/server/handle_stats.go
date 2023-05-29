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
	userStat, err := s.db.GetUserStats(context.Background(), userIdInt)
	if err != nil {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, err.Error())
		return
	}
	writeJsonResponse(w, r.Method, http.StatusOK, userStat)
}
