package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/DennisPing/cs6650-twinder-a3/lib/models"
	"github.com/go-chi/chi"
)

// POST /swipe/{leftorright}/
func (s *Server) PostSwipe(w http.ResponseWriter, r *http.Request) {
	leftorright := chi.URLParam(r, "leftorright")

	var sr models.SwipeRequest
	err := json.NewDecoder(r.Body).Decode(&sr)
	if err != nil {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, "bad request")
		return
	}

	if _, err := strconv.Atoi(sr.Swiper); err != nil {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, fmt.Sprintf("invalid swiper: %s", sr.Swiper))
		return
	}
	if _, err := strconv.Atoi(sr.Swipee); err != nil {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, fmt.Sprintf("invalid swipee: %s", sr.Swipee))
		return
	}
	if len(sr.Comment) > 256 {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, "comment too long")
		return
	}
	if leftorright != "left" && leftorright != "right" {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, fmt.Sprintf("not left or right: %s", leftorright))
		return
	}

	// Append the direction to the request body
	sr.Direction = leftorright

	// Publish the message
	if err = s.PublishToRmq(sr); err != nil {
		writeErrorResponse(w, r.Method, http.StatusInternalServerError, "failed to publish message")
		return
	}

	// Left and right do the same thing for now
	// Always return a response back to client since this is asynchronous, don't let them know about RabbitMQ
	switch leftorright {
	case "left":
		writeStatusResponse(w, r.Method, http.StatusCreated)
		s.metrics.IncrementThroughput()
	case "right":
		writeStatusResponse(w, r.Method, http.StatusCreated)
		s.metrics.IncrementThroughput()
	}
}
