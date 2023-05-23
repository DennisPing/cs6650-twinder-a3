package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/DennisPing/cs6650-twinder-a3/httpserver/metrics"
	"github.com/DennisPing/cs6650-twinder-a3/httpserver/rmqproducer"
	"github.com/DennisPing/cs6650-twinder-a3/lib/logger"
	"github.com/DennisPing/cs6650-twinder-a3/lib/models"
	"github.com/go-chi/chi"
	"github.com/wagslane/go-rabbitmq"
)

type Server struct {
	http.Server
	metrics.Metrics
	rmqproducer.Publisher
	ticker      *time.Ticker
	cancelToken context.CancelFunc
}

// Create a new server which is composed of an HTTP server and a RabbitMQ publisher
func NewServer(addr string, metrics metrics.Metrics, publisher rmqproducer.Publisher) *Server {
	chiRouter := chi.NewRouter()
	s := &Server{
		Server: http.Server{
			Addr:    addr,
			Handler: chiRouter,
		},
		Metrics:   metrics,
		Publisher: publisher,
	}
	chiRouter.Get("/health", s.homeHandler)
	chiRouter.Post("/swipe/{leftorright}/", s.swipeHandler)
	return s
}

// Start the server and start metrics on a new goroutine
func (s *Server) Start() error {
	s.ticker = time.NewTicker(5 * time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelToken = cancel
	go func() { // Metrics goroutine
		for {
			select {
			case <-ctx.Done(): // Quit
				return
			case <-s.ticker.C: // Keep on ticking
				err := s.SendMetrics()
				if err != nil {
					logger.Error().Msgf("unable to send metrics to Axiom: %v", err)
				}
			}
		}
	}()
	return s.ListenAndServe()
}

// Stop the server and stop the ticker
func (s *Server) Stop() {
	s.cancelToken() // Stop the metrics goroutine
	s.ticker.Stop() // Stop the ticker

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(shutdownCtx); err != nil {
		logger.Error().Msgf("Failed to shutdown HTTP server gracefully: %v", err)
	}
}

// Health endpoint
func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello world!"))
}

// Handle swipe left or right
func (s *Server) swipeHandler(w http.ResponseWriter, r *http.Request) {
	leftorright := chi.URLParam(r, "leftorright")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, "bad request")
		return
	}

	var reqBody models.SwipePayload
	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, "bad request")
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

	// Left and right do the same thing for now
	// Always return a response back to client since this is asynchronous, don't let them know about RabbitMQ
	switch leftorright {
	case "left":
		writeStatusResponse(w, http.StatusCreated)
		s.IncrementThroughput()
	case "right":
		writeStatusResponse(w, http.StatusCreated)
		s.IncrementThroughput()
	default:
		writeErrorResponse(w, r.Method, http.StatusBadRequest, "not left or right")
		return
	}

	// Append the direction to the request body
	reqBody.Direction = leftorright

	// Publish the message
	if err = s.PublishToRmq(reqBody); err != nil {
		logger.Error().Msgf("failed to publish to rabbitmq: %v", err)
	}
}

// Publish a message out to the RabbitMQ exchange
func (s *Server) PublishToRmq(payload interface{}) error {
	logger.Debug().Interface("message", payload).Send()
	respBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.Publish(
		[]byte(respBytes),
		[]string{""},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsExchange("swipes"),
	)
}

// Write a simple HTTP status to the response writer
func writeStatusResponse(w http.ResponseWriter, statusCode int) {
	logger.Debug().Int("code", statusCode)
	w.WriteHeader(statusCode)
}

// Marshal and write a JSON response to the response writer
func writeJsonResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	respBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Error().Int("code", http.StatusInternalServerError).Msg("error marshaling JSON response")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	logger.Debug().Interface("send", payload).Send()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(respBytes)))
	w.WriteHeader(statusCode)
	w.Write(respBytes)
}

// Write an HTTP error to the response writer
func writeErrorResponse(w http.ResponseWriter, method string, statusCode int, message string) {
	logger.Warn().Str("method", method).Int("code", statusCode).Msg(message)
	http.Error(w, message, statusCode)
}
