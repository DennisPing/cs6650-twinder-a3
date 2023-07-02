package rmqconsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/DennisPing/cs6650-twinder-a3/consumer/store"
	"github.com/DennisPing/cs6650-twinder-a3/lib/logger"
	"github.com/DennisPing/cs6650-twinder-a3/lib/models"
	"github.com/wagslane/go-rabbitmq"
)

var zlog = logger.GetLogger()

// A rabbitmq consumer + dynamodb client
type ConsumerClient struct {
	Conn     *rabbitmq.Conn
	Consumer *rabbitmq.Consumer
	Store    *store.DatabaseClient
}

func NewConsumerClient(conn *rabbitmq.Conn, store *store.DatabaseClient) (*ConsumerClient, error) {
	cc := &ConsumerClient{
		Conn:  conn,
		Store: store,
	}
	consumer, err := rabbitmq.NewConsumer(
		conn,
		cc.HandleMessage,
		"",
		rabbitmq.WithConsumerOptionsLogging,
		rabbitmq.WithConsumerOptionsRoutingKey(""), // Bind this default queue to default routing key
		rabbitmq.WithConsumerOptionsExchangeDeclare,
		rabbitmq.WithConsumerOptionsExchangeName("swipes"),
		rabbitmq.WithConsumerOptionsExchangeKind("fanout"),
		rabbitmq.WithConsumerOptionsQOSPrefetch(64),
		rabbitmq.WithConsumerOptionsConcurrency(50),
		rabbitmq.WithConsumerOptionsQueueAutoDelete, // Auto delete the queue upon disconnect
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create rabbitmq consumer: %w", err)
	}
	cc.Consumer = consumer
	return cc, nil
}

func (cc *ConsumerClient) HandleMessage(d rabbitmq.Delivery) rabbitmq.Action {
	zlog.Debug().Msg(string(d.Body))

	var reqBody models.SwipeRequest
	err := json.Unmarshal(d.Body, &reqBody)
	if err != nil {
		zlog.Error().Err(err).Msg("bad request")
		return rabbitmq.NackDiscard
	}

	userId, _ := strconv.Atoi(reqBody.Swiper)
	swipee, _ := strconv.Atoi(reqBody.Swipee)
	err = cc.Store.UpdateUserStats(context.Background(), userId, swipee, reqBody.Direction)
	if err != nil {
		zlog.Error().Err(err).Interface("payload", reqBody).Msg("consumer failed on UpdateUserStats")
	}
	return rabbitmq.Ack
}

// Close the rabbitmq consumer and the underlying TCP connection
func (cc *ConsumerClient) Close() {
	cc.Consumer.Close()
	cc.Conn.Close()
}

// Init a new RabbitMQ connection with the RabbitMQ host.
func NewRmqConn() (*rabbitmq.Conn, error) {
	host := os.Getenv("RABBITMQ_HOST")

	if host == "" {
		zlog.Fatal().Msg("you forgot to set the RABBITMQ_HOST env variable")
	}

	// Create a new connection to rabbitmq
	return rabbitmq.NewConn(
		fmt.Sprintf("amqp://%s:%s@%s:5672", "guest", "guest", host),
		rabbitmq.WithConnectionOptionsLogging,
	)
}
