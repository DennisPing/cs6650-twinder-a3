package db

import (
	"context"
	"errors"
	"testing"

	"github.com/DennisPing/cs6650-twinder-a3/httpserver/db/mocks"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Get user stats happy path
func TestGetUserStats(t *testing.T) {
	ctx := context.Background()
	mockDynamo := mocks.NewDynamoClienter(t)

	mockDynamo.EXPECT().
		GetItem(ctx, mock.Anything, mock.Anything).
		Return(&dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"numLikes":    &types.AttributeValueMemberN{Value: "11"},
				"numDislikes": &types.AttributeValueMemberN{Value: "22"},
			},
		}, nil)

	databaseClient := DatabaseClient{
		Client: mockDynamo,
		Table:  "testTable",
	}

	stats, err := databaseClient.GetUserStats(ctx, 1234)

	mockDynamo.AssertExpectations(t)

	assert.NoError(t, err)
	assert.Equal(t, 11, stats.NumLikes)
	assert.Equal(t, 22, stats.NumDislikes)
}

// Get user stats sad path
func TestGetUserStatsError(t *testing.T) {
	userId := 1234
	tests := []struct {
		name              string
		mockItem          *dynamodb.GetItemOutput
		mockInternalError error
		expectedErrorMsg  string // the high level error message
	}{
		{
			name:              "dynamo internal error",
			mockItem:          nil,
			mockInternalError: errors.New("aws died"),
			expectedErrorMsg:  "failed to get item: aws died",
		},
		{
			name:              "no error but item not found",
			mockItem:          &dynamodb.GetItemOutput{},
			mockInternalError: nil,
			expectedErrorMsg:  "userId not found: 1234",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			mockDynamo := mocks.NewDynamoClienter(t)

			mockDynamo.EXPECT().
				GetItem(ctx, mock.Anything, mock.Anything).
				Return(tc.mockItem, tc.mockInternalError)

			databaseClient := DatabaseClient{
				Client: mockDynamo,
				Table:  "testTable",
			}

			_, err := databaseClient.GetUserStats(ctx, userId)

			mockDynamo.AssertExpectations(t)

			assert.Error(t, err) // check for a high level error
			assert.Equal(t, tc.expectedErrorMsg, err.Error())

			// Everything works, but user not found
			if tc.mockInternalError == nil {
				assert.Nil(t, tc.mockItem.Item)
			}
		})
	}
}

// Update user stats happy path
func TestUpdateUserStats(t *testing.T) {
	tests := []struct {
		name     string
		userId   int
		swipee   int
		swipeDir string
	}{
		{
			name:     "swipe right",
			userId:   1234,
			swipee:   5678,
			swipeDir: "right",
		},
		{
			name:     "swipe left",
			userId:   1234,
			swipee:   5678,
			swipeDir: "left",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			mockDynamoClient := mocks.NewDynamoClienter(t)

			mockDynamoClient.EXPECT().
				UpdateItem(ctx, mock.Anything, mock.Anything).
				Return(&dynamodb.UpdateItemOutput{}, nil)

			databaseClient := DatabaseClient{
				Client: mockDynamoClient,
				Table:  "testTable",
			}

			err := databaseClient.UpdateUserStats(ctx, tc.userId, tc.swipee, tc.swipeDir)

			mockDynamoClient.AssertExpectations(t)

			assert.NoError(t, err)
		})
	}
}

// Update user stats sad path
func TestUpdateUserStatsError(t *testing.T) {
	tests := []struct {
		name              string
		userId            int
		swipee            int
		swipeDir          string
		mockInternalError error
		expectedErrorMsg  string // the high level error message
	}{
		{
			name:              "dynamo internal error",
			userId:            1234,
			swipee:            5678,
			swipeDir:          "right",
			mockInternalError: errors.New("aws died"),
			expectedErrorMsg:  "aws died",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			mockDynamoClient := mocks.NewDynamoClienter(t)

			mockDynamoClient.EXPECT().
				UpdateItem(ctx, mock.Anything, mock.Anything).
				Return(&dynamodb.UpdateItemOutput{}, tc.mockInternalError)

			databaseClient := DatabaseClient{
				Client: mockDynamoClient,
				Table:  "testTable",
			}

			err := databaseClient.UpdateUserStats(ctx, tc.userId, tc.swipee, tc.swipeDir)

			mockDynamoClient.AssertExpectations(t)

			assert.Error(t, err)
			assert.Equal(t, tc.expectedErrorMsg, err.Error())
		})
	}
}
