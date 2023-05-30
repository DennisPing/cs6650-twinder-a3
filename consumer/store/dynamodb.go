package store

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DennisPing/cs6650-twinder-a3/lib/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
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
	Table  string
}

// Create a new DatabaseClient that has a DynamoDB client
func NewDatabaseClient() (*DatabaseClient, error) {
	// LoadDefaultConfig automatically loads env variables and ~/.aws/credentials
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion("us-east-2"),
		config.WithRetryer(func() aws.Retryer {
			return aws.NopRetryer{} // Don't retry errors
		}),
	)
	if err != nil {
		return nil, err
	}
	return &DatabaseClient{
		Client: dynamodb.NewFromConfig(cfg),
		Table:  "SwipeData",
	}, nil
}

// Update a user's stats. If userId doesn't exist, then a new entry is created
func (d *DatabaseClient) UpdateUserStats(ctx context.Context, userId, swipee int, swipeDir string) error {
	logger.Debug().Msg("Hello from UpdateUserStats")
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
	logger.Debug().Msg("Built expression")

	// Execute the UpdateItem operation
	_, err = d.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberN{Value: strconv.Itoa(userId)},
		},
		TableName:                 aws.String(d.Table),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueNone,
	})
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}
	logger.Debug().Msg("Completed UpdateUserStats")
	return nil
}
