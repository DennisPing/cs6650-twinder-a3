package dynamodb_repo

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// The low level DynamoDB service
//
//go:generate mockery --name=DynamoDBServicer --filename=mock_service.go
type DynamoDBServicer interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

type DynamoDBService struct {
	// Having this implies that we want to make real network calls
	client *dynamodb.Client
}

// Create a new low level DynamoDB service
func NewDynamoDBService(client *dynamodb.Client) *DynamoDBService {
	return &DynamoDBService{
		client: client,
	}
}
