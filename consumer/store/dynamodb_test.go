package store

import (
	"context"
	"errors"
	"testing"

	"github.com/DennisPing/cs6650-twinder-a3/consumer/store/mocks"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
			}

			err := databaseClient.UpdateUserStats(ctx, tc.userId, tc.swipee, tc.swipeDir)

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
			expectedErrorMsg:  "UpdateItem failed: aws died",
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
			}

			err := databaseClient.UpdateUserStats(ctx, tc.userId, tc.swipee, tc.swipeDir)

			assert.Error(t, err)
			assert.Equal(t, tc.expectedErrorMsg, err.Error())
		})
	}
}
