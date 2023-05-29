package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/DennisPing/cs6650-twinder-a3/httpserver/db"
	"github.com/DennisPing/cs6650-twinder-a3/httpserver/metrics"
	"github.com/DennisPing/cs6650-twinder-a3/httpserver/rmqproducer"
	"github.com/DennisPing/cs6650-twinder-a3/lib/logger"
	"github.com/DennisPing/cs6650-twinder-a3/lib/models"
	"github.com/go-chi/chi"
	"github.com/wagslane/go-rabbitmq"
)

type Server struct {
	http.Server
	metrics metrics.Metrics       // interface
	pub     rmqproducer.Publisher // interface
	db      *db.DatabaseClient
	ticker  *time.Ticker
	cancel  context.CancelFunc
}

// Create a new server which has an HTTP server, Metrics client, RabbitMQ publisher, and Database client
func NewServer(addr string, metrics metrics.Metrics, publisher rmqproducer.Publisher, dbClient *db.DatabaseClient) *Server {
	chiRouter := chi.NewRouter()

	// Build the server
	s := &Server{
		Server: http.Server{
			Addr:    addr,
			Handler: chiRouter,
		},
		metrics: metrics,
		pub:     publisher,
		db:      dbClient,
	}
	chiRouter.Get("/health", s.HomeHandler)
	chiRouter.Post("/swipe/{leftorright}/", s.SwipeHandler)
	chiRouter.Get("/matches/{userId}/", s.MatchesHandler)
	chiRouter.Get("/stats/{userId}/", s.StatsHandler)
	return s
}

// Start the server and start metrics on a new goroutine
func (s *Server) Start() error {
	s.ticker = time.NewTicker(5 * time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go func() { // Metrics goroutine
		for {
			select {
			case <-ctx.Done(): // Quit
				return
			case <-s.ticker.C: // Keep on ticking
				err := s.metrics.SendMetrics()
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
	s.cancel()      // Stop the metrics goroutine
	s.ticker.Stop() // Stop the ticker

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(shutdownCtx); err != nil {
		logger.Error().Msgf("Failed to shutdown HTTP server gracefully: %v", err)
	}
}

// Publish a message out to the RabbitMQ exchange
func (s *Server) PublishToRmq(payload interface{}) error {
	logger.Debug().Interface("message", payload).Send()
	respBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.pub.Publish(
		[]byte(respBytes),
		[]string{""},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsExchange("swipes"),
	)
}

// Send a simple HTTP response with no payload
func writeStatusResponse(w http.ResponseWriter, method string, statusCode int) {
	logger.Debug().Str("method", method).Int("code", statusCode)
	w.WriteHeader(statusCode)
}

// Send an HTTP response with JSON payload
func writeJsonResponse(w http.ResponseWriter, method string, statusCode int, payload interface{}) {
	logger.Debug().Str("method", method).Interface("payload", payload).Send()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		writeErrorResponse(w, method, http.StatusInternalServerError, "failed to encode JSON response")
	}
}

// Send an HTTP response error with a message
func writeErrorResponse(w http.ResponseWriter, method string, statusCode int, message string) {
	logger.Warn().Str("method", method).Int("code", statusCode).Msg(message)
	errorResp := &models.ErrorResponse{
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(errorResp)
}
