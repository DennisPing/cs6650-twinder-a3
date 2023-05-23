package rmqconsumer

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/DennisPing/cs6650-twinder-a3/consumer/store"
	"github.com/DennisPing/cs6650-twinder-a3/lib/logger"
	"github.com/DennisPing/cs6650-twinder-a3/lib/models"
	"github.com/wagslane/go-rabbitmq"
)

// Init a new RabbitMQ connection with the RabbitMQ host.
func NewRmqConn() (*rabbitmq.Conn, error) {
	host := os.Getenv("RABBITMQ_HOST")

	if host == "" {
		logger.Fatal().Msg("you forgot to set the RABBITMQ_HOST environment variable")
	}

	// Create a new connection to rabbitmq
	return rabbitmq.NewConn(
		fmt.Sprintf("amqp://%s:%s@%s:5672", "guest", "guest", host),
		rabbitmq.WithConnectionOptionsLogging,
	)
}

// Start the RabbitMQ consumer. It parses the message and adds a new UserStat into the kv store.
func StartRmqConsumer(conn *rabbitmq.Conn, kvStore *store.SimpleStore) (*rabbitmq.Consumer, error) {
	return rabbitmq.NewConsumer(
		conn,
		func(d rabbitmq.Delivery) rabbitmq.Action {
			logger.Debug().Msg(string(d.Body))

			var reqBody models.SwipePayload
			err := json.Unmarshal(d.Body, &reqBody)
			if err != nil {
				logger.Error().Msgf("bad request: %v", err)
				return rabbitmq.NackDiscard
			}

			kvStore.Add(reqBody.Swiper, reqBody.Swipee, reqBody.Direction)
			return rabbitmq.Ack
		},
		"",
		rabbitmq.WithConsumerOptionsLogging,
		rabbitmq.WithConsumerOptionsRoutingKey(""), // Bind this default queue to default routing key
		rabbitmq.WithConsumerOptionsExchangeDeclare,
		rabbitmq.WithConsumerOptionsExchangeName("swipes"),
		rabbitmq.WithConsumerOptionsExchangeKind("fanout"),
		rabbitmq.WithConsumerOptionsQOSPrefetch(128),
		rabbitmq.WithConsumerOptionsConcurrency(4),
		rabbitmq.WithConsumerOptionsQueueAutoDelete, // Auto delete the queue upon disconnect
	)
}
