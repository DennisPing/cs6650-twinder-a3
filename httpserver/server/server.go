package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/DennisPing/cs6650-twinder-a3/httpserver/db"
	"github.com/DennisPing/cs6650-twinder-a3/httpserver/metrics"
	"github.com/DennisPing/cs6650-twinder-a3/httpserver/rmqproducer"
	"github.com/DennisPing/cs6650-twinder-a3/lib/logger"
	"github.com/go-chi/chi"
	"github.com/wagslane/go-rabbitmq"
)

type Server struct {
	http.Server
	metrics  metrics.Metrics       // interface
	pub      rmqproducer.Publisher // interface
	database db.MongoDB            // interface
	ticker   *time.Ticker
	cancel   context.CancelFunc
}

// Create a new server which is composed of an HTTP server, RabbitMQ publisher, and MongoDB client
func NewServer(addr string, metrics metrics.Metrics, publisher rmqproducer.Publisher, mongoClient db.MongoDB) *Server {
	chiRouter := chi.NewRouter()

	// Build the server
	s := &Server{
		Server: http.Server{
			Addr:    addr,
			Handler: chiRouter,
		},
		metrics:  metrics,
		pub:      publisher,
		database: mongoClient,
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

	err := s.database.Disconnect(context.Background())
	if err != nil {
		logger.Error().Msgf("Failed to disconnect MongoDB client: %v", err)
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
