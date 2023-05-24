package db

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//go:generate mockery --name=MongoDB --filename=mock_mongo.go
type MongoDB interface {
	Connect(ctx context.Context) error
	Ping(ctx context.Context) error
	Database(name string) *mongo.Database
	Disconnect(ctx context.Context) error
}

type MongoDBClient struct {
	Client *mongo.Client
}

func NewMongoDBClient(uri string) (*MongoDBClient, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}
	return &MongoDBClient{Client: client}, nil
}

func (m *MongoDBClient) Connect(ctx context.Context) error {
	return m.Client.Connect(ctx)
}

func (m *MongoDBClient) Ping(ctx context.Context) error {
	return m.Client.Ping(ctx, nil)
}

func (m *MongoDBClient) Database(name string) *mongo.Database {
	return m.Client.Database(name)
}

func (m *MongoDBClient) Disconnect(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}
