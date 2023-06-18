package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/DennisPing/cs6650-twinder-a3/httpserver/metrics"
	"github.com/DennisPing/cs6650-twinder-a3/httpserver/middleware"
	"github.com/DennisPing/cs6650-twinder-a3/httpserver/rmqproducer"
	"github.com/DennisPing/cs6650-twinder-a3/httpserver/store"
	"github.com/DennisPing/cs6650-twinder-a3/lib/logger"
	"github.com/DennisPing/cs6650-twinder-a3/lib/models"
	"github.com/go-chi/chi"
	"github.com/wagslane/go-rabbitmq"
)

var zlog = logger.GetLogger()

type Server struct {
	http.Server
	metrics metrics.Metrics       // interface
	pub     rmqproducer.Publisher // interface
	store   *store.DatabaseClient
	ticker  *time.Ticker
	cancel  context.CancelFunc
}

// Create a new server which has an HTTP server, Metrics client, RabbitMQ publisher, and Database client
func NewServer(addr string, metrics metrics.Metrics, publisher rmqproducer.Publisher, dbClient *store.DatabaseClient) *Server {
	chiRouter := chi.NewRouter()
	chiRouter.Use(middleware.LoggingMiddleware)

	// Build the server
	s := &Server{
		Server: http.Server{
			Addr:    addr,
			Handler: chiRouter,
		},
		metrics: metrics,
		pub:     publisher,
		store:   dbClient,
	}
	chiRouter.Get("/health", s.GetHealth)
	chiRouter.Post("/swipe/{leftorright}/", s.PostSwipe)
	chiRouter.Get("/matches/{userId}/", s.GetMatches)
	chiRouter.Get("/stats/{userId}/", s.GetStats)
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
					zlog.Error().Err(err).Msg("unable to send metrics to Axiom")
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
		zlog.Error().Err(err).Msg("failed to shutdown HTTP server gracefully")
	}
}

// Publish a message out to the RabbitMQ exchange
func (s *Server) PublishToRmq(payload interface{}) error {
	zlog.Debug().Interface("payload", payload).Msg("publish")

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
	zlog.Debug().Str("method", method).Int("code", statusCode).Msg("response")
	w.WriteHeader(statusCode)
}

// Send an HTTP response with JSON payload
func writeJsonResponse(w http.ResponseWriter, method string, statusCode int, payload interface{}) {
	zlog.Debug().Str("method", method).Interface("payload", payload).Msg("response")
	respBytes, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(respBytes)))
	w.WriteHeader(statusCode)
	w.Write(respBytes)
}

// Send an HTTP response error with a message
func writeErrorResponse(w http.ResponseWriter, method string, statusCode int, message string) {
	zlog.Warn().Str("method", method).Int("code", statusCode).Msg(message)
	errBytes, err := json.Marshal(
		&models.ErrorResponse{
			Message: message,
		})
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(errBytes)))
	w.WriteHeader(statusCode)
	w.Write(errBytes)
}
