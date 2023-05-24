package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/DennisPing/cs6650-twinder-a3/lib/models"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GET /stats/{userId}/
func (s *Server) StatsHandler(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, fmt.Sprintf("invalid userId: %s", userId))
		return
	}

	var result models.UserStats
	err = s.database.Database("twinder").Collection("swipedata").FindOne(context.Background(), bson.M{"_id": userIdInt}).Decode(&result)
	if err != nil {
		// if the document is not found, MongoDB returns a mongo.ErrNoDocuments error
		if err == mongo.ErrNoDocuments {
			writeErrorResponse(w, r.Method, http.StatusNotFound, fmt.Sprintf("userId not found: %s", userId))
		} else {
			writeErrorResponse(w, r.Method, http.StatusInternalServerError, fmt.Sprintf("error occurred while fetching user stats: %v", err))
		}
		return
	}

	// Process and return data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
