package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/DennisPing/cs6650-twinder-a3/consumer/rmqconsumer"
	"github.com/DennisPing/cs6650-twinder-a3/consumer/store"
	"github.com/DennisPing/cs6650-twinder-a3/lib/logger"
)

func main() {
	store, err := store.NewDatabaseClient()
	if err != nil {
		logger.Fatal().Err(err).Msg("unable to connect to DynamoDB")
	}

	rmqConn, err := rmqconsumer.NewRmqConn()
	if err != nil {
		logger.Fatal().Err(err).Msg("unable to make rabbitmq connection")
	}
	defer rmqConn.Close()

	consumer, err := rmqconsumer.StartRmqConsumer(rmqConn, store)
	if err != nil {
		logger.Fatal().Err(err)
	}
	defer consumer.Close()

	// Set up a signal handler for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("shutting down gracefully...")
}
