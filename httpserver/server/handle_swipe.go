package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/DennisPing/cs6650-twinder-a3/lib/models"
	"github.com/go-chi/chi"
)

// POST /swipe/{leftorright}/
func (s *Server) PostSwipe(w http.ResponseWriter, r *http.Request) {
	leftorright := chi.URLParam(r, "leftorright")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, "bad request")
		return
	}

	var reqBody models.SwipeRequest
	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, "corrupt request body")
		return
	}
	if _, err := strconv.Atoi(reqBody.Swiper); err != nil {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, fmt.Sprintf("invalid swiper: %s", reqBody.Swiper))
		return
	}
	if _, err := strconv.Atoi(reqBody.Swipee); err != nil {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, fmt.Sprintf("invalid swipee: %s", reqBody.Swipee))
		return
	}
	if len(reqBody.Comment) > 256 {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, "comment too long")
		return
	}
	if leftorright != "left" && leftorright != "right" {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, fmt.Sprintf("not left or right: %s", leftorright))
		return
	}

	// Append the direction to the request body
	reqBody.Direction = leftorright

	// Publish the message
	if err = s.PublishToRmq(reqBody); err != nil {
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
