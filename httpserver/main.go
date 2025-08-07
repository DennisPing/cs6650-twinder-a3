package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/DennisPing/cs6650-twinder-a3/httpserver/db/dynamodb_repo"
	"github.com/DennisPing/cs6650-twinder-a3/httpserver/metrics"
	"github.com/DennisPing/cs6650-twinder-a3/httpserver/rmqproducer"
	"github.com/DennisPing/cs6650-twinder-a3/httpserver/server"
	"github.com/DennisPing/cs6650-twinder-a3/lib/logger"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	addr := fmt.Sprintf(":%s", port)

	log := logger.InitLogger()

	// Initialize metrics client
	metricsClient, err := metrics.NewMetricsClient()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to set up metrics")
	}

	// Initialize rabbitmq publisher
	rmqConn, err := rmqproducer.NewConnection()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to make rabbitmq connection")
	}
	defer rmqConn.Close()
	publisher, err := rmqproducer.NewPublisher(rmqConn)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to make rabbitmq publisher")
	}
	defer publisher.Close()

	// Initialize database client
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion("us-east-2"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to load dynamodb config")
	}

	dbClient := dynamodb_repo.NewDynamoDBRepo(dynamodb.NewFromConfig(cfg))
	if err != nil {
		log.Fatal().Err(err).Msg("unable to connect to DynamoDB")
	}
	log.Info().Msg("connected to DynamoDB")

	// Initialize the http server
	server := server.NewServer(addr, log, metricsClient, publisher, dbClient)

	// Run the http server in a goroutine
	fmt.Printf("Starting server on port %s...\n", port)
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server died")
		}
	}()

	// Set up a signal handler for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until quit signal
	<-quit
	log.Info().Msg("Shutting down gracefully...")
	server.Stop()
}
