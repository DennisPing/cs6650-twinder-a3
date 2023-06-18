package store

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

//go:generate mockery --name=DynamoClienter --filename=mock_database.go
type DynamoClienter interface {
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

// Database client that communicates with DynamoDB via interface
type DatabaseClient struct {
	Client DynamoClienter // Interface
}

// Create a new DatabaseClient that has a DynamoDB client
func NewDatabaseClient() (*DatabaseClient, error) {
	// LoadDefaultConfig automatically loads env variables and ~/.aws/credentials
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion("us-east-2"),
	)
	if err != nil {
		return nil, err
	}
	return &DatabaseClient{
		Client: dynamodb.NewFromConfig(cfg),
	}, nil
}

// Update a user's stats. If userId doesn't exist, then a new entry is created
func (d *DatabaseClient) UpdateUserStats(ctx context.Context, userId, swipee int, swipeDir string) error {
	tableName := getTableShard(userId)
	var update expression.UpdateBuilder
	switch swipeDir {
	case "right":
		update = update.
			Add(expression.Name("numLikes"), expression.Value(1)).
			Set(expression.Name("numDislikes"), expression.IfNotExists(expression.Name("numDislikes"), expression.Value(0))).
			Add(expression.Name("matchList"), expression.Value(
				&types.AttributeValueMemberNS{Value: []string{strconv.Itoa(swipee)}},
			))
	case "left":
		update = update.
			Add(expression.Name("numDislikes"), expression.Value(1)).
			Set(expression.Name("numLikes"), expression.IfNotExists(expression.Name("numLikes"), expression.Value(0)))
	default:
		return fmt.Errorf("invalid swipe direction: %s", swipeDir)
	}

	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return fmt.Errorf("failed to build expression: %w", err)
	}

	// Execute the UpdateItem operation
	_, err = d.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberN{Value: strconv.Itoa(userId)},
		},
		TableName:                 &tableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueNone,
	})
	if err != nil {
		return fmt.Errorf("UpdateItem failed: %w", err)
	}

	// if err != nil {
	// 	// https://github.com/golang/go/issues/37625#issuecomment-594033043
	// 	var opErr *smithy.OperationError
	// 	if errors.As(err, &opErr) {
	// 		switch opErr.Unwrap().(type) {
	// 		case *types.ProvisionedThroughputExceededException:
	// 			return fmt.Errorf("update rate too high: %w", err)
	// 		default:
	// 			return fmt.Errorf("operation error: %w", err)
	// 		}
	// 	}
	// 	return fmt.Errorf("unexpected error: %w", err)
	// }
	return nil
}
