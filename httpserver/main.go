package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/DennisPing/cs6650-twinder-a3/httpserver/metrics"
	"github.com/DennisPing/cs6650-twinder-a3/httpserver/rmqproducer"
	"github.com/DennisPing/cs6650-twinder-a3/httpserver/server"
	"github.com/DennisPing/cs6650-twinder-a3/lib/logger"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)

	// Initialize metrics client
	metrics, err := metrics.NewMetrics()
	if err != nil {
		logger.Fatal().Msgf("unable to set up metrics: %v", err)
	}

	// Initialize rabbitmq publisher
	rmqConn, err := rmqproducer.NewConnection()
	if err != nil {
		logger.Fatal().Msgf("unable to make rabbitmq connection: %v", err)
	}
	defer rmqConn.Close()
	publisher, err := rmqproducer.NewPublisher(rmqConn)
	if err != nil {
		logger.Fatal().Msgf("unable to make rabbitmq publisher: %v", err)
	}
	defer publisher.Close()

	// Initialize the http server
	server := server.NewServer(addr, metrics, publisher)

	// Run the http server in a goroutine
	fmt.Printf("Starting server on port %s...\n", port)
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Msgf("server died: %v", err)
		}
	}()

	// Set up a signal handler for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until quit signal
	<-quit
	logger.Info().Msg("Shutting down gracefully...")
	server.Stop()
}
