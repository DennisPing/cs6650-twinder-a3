package rmqproducer

import (
	"errors"
	"fmt"
	"os"

	"github.com/wagslane/go-rabbitmq"
)

//go:generate mockery --name=Publisher --filename=mock_publisher.go
type Publisher interface {
	Publish(data []byte, routingKeys []string, optionFuncs ...func(*rabbitmq.PublishOptions)) error
}

// Init a new RabbitMQ connection with the RabbitMQ host.
func NewConnection() (*rabbitmq.Conn, error) {
	host := os.Getenv("RABBITMQ_HOST")

	if host == "" {
		return nil, errors.New("you forgot to set the RABBITMQ_HOST environment variable")
	}

	conn, err := rabbitmq.NewConn(
		fmt.Sprintf("amqp://%s:%s@%s:5672", "guest", "guest", host),
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Create a new publisher that publishes to the "swipes" exchange via "fanout" method.
func NewPublisher(conn *rabbitmq.Conn) (*rabbitmq.Publisher, error) {
	publisher, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsExchangeDeclare,
		rabbitmq.WithPublisherOptionsExchangeName("swipes"),
		rabbitmq.WithPublisherOptionsExchangeKind("fanout"),
	)
	if err != nil {
		return nil, err
	}
	return publisher, nil
}
