package db

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DennisPing/cs6650-twinder-a3/lib/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

//go:generate mockery --name=DynamoClienter --filename=mock_database.go
type DynamoClienter interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
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
			return aws.NopRetryer{}
		}),
	)
	if err != nil {
		return nil, err
	}
	client := dynamodb.NewFromConfig(cfg)
	return &DatabaseClient{
		Client: client,
		Table:  "SwipeData",
	}, nil
}

// Get the user stats of a userId
func (d *DatabaseClient) GetUserStats(ctx context.Context, userId int) (*models.UserStats, error) {
	item, err := d.getItem(ctx, userId)
	if err != nil {
		return nil, err
	} else if item != nil {
		return &models.UserStats{
			NumLikes:    item.NumLikes,
			NumDislikes: item.NumDislikes,
		}, nil
	} else {
		return nil, fmt.Errorf("userId not found: %d", userId)
	}
}

func (d *DatabaseClient) GetMatches(ctx context.Context, userId int) (*models.UserMatches, error) {
	item, err := d.getItem(ctx, userId)
	if err != nil {
		return nil, err
	} else if item != nil {
		return &models.UserMatches{
			MatchList: item.MatchList,
		}, nil
	} else {
		return nil, fmt.Errorf("userId not found: %d", userId)
	}
}

// Update a user's stats. If userId doesn't exist, then a new entry is created
func (d *DatabaseClient) UpdateUserStats(ctx context.Context, userId, swipee int, swipeDir string) error {
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
		TableName:                 aws.String(d.Table),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueNone,
	})
	return err
}

// Helper method to get DynamoDB item
func (d *DatabaseClient) getItem(ctx context.Context, userId int) (*models.DynamoUserStats, error) {
	resp, err := d.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.Table),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberN{Value: strconv.Itoa(userId)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}
	if resp.Item == nil {
		return nil, nil
	}
	dynamoItem := &models.DynamoUserStats{}
	if err = attributevalue.UnmarshalMap(resp.Item, &dynamoItem); err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %w", err)
	}
	// logger.Debug().Str("method", "GET").Interface("dynamoItem", dynamoItem).Send()
	return dynamoItem, nil
}
