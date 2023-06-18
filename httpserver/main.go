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
	"github.com/DennisPing/cs6650-twinder-a3/httpserver/store"
	"github.com/DennisPing/cs6650-twinder-a3/lib/logger"
)

var zlog = logger.GetLogger()

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	addr := fmt.Sprintf(":%s", port)

	// Initialize metrics client
	metricsClient, err := metrics.NewMetricsClient()
	if err != nil {
		zlog.Fatal().Err(err).Msg("unable to set up metrics")
	}

	// Initialize rabbitmq publisher
	rmqConn, err := rmqproducer.NewConnection()
	if err != nil {
		zlog.Fatal().Err(err).Msg("unable to make rabbitmq connection")
	}
	defer rmqConn.Close()
	publisher, err := rmqproducer.NewPublisher(rmqConn)
	if err != nil {
		zlog.Fatal().Err(err).Msg("unable to make rabbitmq publisher")
	}
	defer publisher.Close()

	// Initialize database client
	dbClient, err := store.NewDatabaseClient()
	if err != nil {
		zlog.Fatal().Err(err).Msg("unable to connect to DynamoDB")
	}
	zlog.Info().Msg("connected to DynamoDB")

	// Initialize the http server
	server := server.NewServer(addr, metricsClient, publisher, dbClient)

	// Run the http server in a goroutine
	fmt.Printf("Starting server on port %s...\n", port)
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			zlog.Fatal().Err(err).Msg("HTTP server died")
		}
	}()

	// Set up a signal handler for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until quit signal
	<-quit
	zlog.Info().Msg("Shutting down gracefully...")
	server.Stop()
}
