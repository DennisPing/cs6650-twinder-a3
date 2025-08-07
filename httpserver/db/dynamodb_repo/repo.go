package dynamodb_repo

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DennisPing/cs6650-twinder-a3/lib/models"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

/**
+---------------+
|  HTTP Server  |
+---------------+
		|
		v
+---------------+
|  Repository   | (our database logic)
+---------------+
		|
		v
+---------------+
|    Service    | (remote network call)
+---------------+
**/

type DynamoDBRepo struct {
	// This can hold a real DynamoDB service or a fake service
	service DynamoDBServicer
}

// Create a new Repository for DynamoDB
func NewDynamoDBRepo(service DynamoDBServicer) *DynamoDBRepo {
	return &DynamoDBRepo{
		service: service,
	}
}

// Get user likes and dislikes. Returns found, user stats, and error.
func (d *DynamoDBRepo) GetUserStats(ctx context.Context, userId int) (models.UserStats, bool, error) {
	var stats models.UserStats

	item, found, err := d.internalGetItem(ctx, userId)
	if err != nil {
		// Some database error
		return stats, false, err
	}
	if !found {
		return stats, false, nil
	}
	stats.NumLikes = item.NumLikes
	stats.NumDislikes = item.NumDislikes

	return stats, true, nil
}

// Get user matches list. Returns found, user matches, and error.
func (d *DynamoDBRepo) GetMatches(ctx context.Context, userId int) (models.UserMatches, bool, error) {
	var matches models.UserMatches

	item, found, err := d.internalGetItem(ctx, userId)
	if err != nil {
		// Some database error
		return matches, false, err
	}
	if !found {
		return matches, false, nil
	}
	matches.MatchList = item.MatchList

	return matches, true, nil
}

// Update a user's stats. If userId doesn't exist, then a new entry is created
func (d *DynamoDBRepo) UpdateUserStats(ctx context.Context, userId, swipee int, swipeDir string) error {
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
	_, err = d.service.UpdateItem(ctx, &dynamodb.UpdateItemInput{
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
		return fmt.Errorf("failed to update item: %w", err)
	}
	return nil
}

// Internal method that gets the entire row from DynamoDB
func (d *DynamoDBRepo) internalGetItem(ctx context.Context, userId int) (models.DynamoUserStats, bool, error) {
	var outItem models.DynamoUserStats

	tableName := getTableShard(userId)
	output, err := d.service.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &tableName,
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberN{Value: strconv.Itoa(userId)},
		},
	})
	if err != nil {
		return outItem, false, fmt.Errorf("failed to get item: %w", err)
	}
	if output.Item == nil {
		// Item not found, but not exactly an error
		return outItem, false, nil
	}
	if err = attributevalue.UnmarshalMap(output.Item, &outItem); err != nil {
		return outItem, false, fmt.Errorf("failed to unmarshal item: %w", err)
	}
	return outItem, true, nil
}
