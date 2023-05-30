package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mockMetrics "github.com/DennisPing/cs6650-twinder-a3/httpserver/metrics/mocks"
	mockPublisher "github.com/DennisPing/cs6650-twinder-a3/httpserver/rmqproducer/mocks"
	"github.com/DennisPing/cs6650-twinder-a3/httpserver/store"
	mockDynamo "github.com/DennisPing/cs6650-twinder-a3/httpserver/store/mocks"
	"github.com/DennisPing/cs6650-twinder-a3/lib/models"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPostSwipe(t *testing.T) {
	// Mock the internal Metrics
	mockMetrics := mockMetrics.NewMetrics(t)
	mockMetrics.EXPECT().IncrementThroughput().Return()

	// Mock the internal Publisher
	mockPublisher := mockPublisher.NewPublisher(t)
	mockPublisher.EXPECT().Publish(
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]string"),
		mock.AnythingOfType("func(*rabbitmq.PublishOptions)"),
		mock.AnythingOfType("func(*rabbitmq.PublishOptions)")).
		Return(nil)

	// Mock the internal Database
	mockDynamoClient := mockDynamo.NewDynamoClienter(t)
	mockDynamoClient.EXPECT().UpdateItem(
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.AnythingOfType("*dynamodb.UpdateItemInput"),
		mock.AnythingOfType("string")).
		Return(nil, nil)

	databaseStub := &store.DatabaseClient{
		Client: mockDynamoClient,
		Table:  "SwipeData",
	}

	s := NewServer(":8080", mockMetrics, mockPublisher, databaseStub)

	tests := []struct {
		name           string
		method         string
		url            string
		body           interface{}
		expectedStatus int
	}{
		{
			name:           "health check",
			method:         "GET",
			url:            "/health",
			body:           nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "swipe left",
			method: "POST",
			url:    "/swipe/left/",
			body: models.SwipeRequest{
				Swiper:  "1234",
				Swipee:  "5678",
				Comment: "asdf"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "swipe right",
			method: "POST",
			url:    "/swipe/right/",
			body: models.SwipeRequest{
				Swiper:  "1234",
				Swipee:  "5678",
				Comment: "asdf"},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tc.body)
			bodyReader := bytes.NewReader(bodyBytes)
			req, _ := http.NewRequest(tc.method, tc.url, bodyReader)

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Serve the request
			s.Handler.ServeHTTP(rr, req)

			resp := rr.Result()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestPostSwipeError(t *testing.T) {
	// Mock internal dependencies
	mockMetrics := mockMetrics.NewMetrics(t)
	mockPublisher := mockPublisher.NewPublisher(t)
	mockDynamoClient := mockDynamo.NewDynamoClienter(t)

	databaseStub := &store.DatabaseClient{
		Client: mockDynamoClient,
		Table:  "SwipeData",
	}

	s := NewServer(":8080", mockMetrics, mockPublisher, databaseStub)

	tests := []struct {
		name            string
		method          string
		url             string
		body            interface{}
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:   "bad swipe direction",
			method: "POST",
			url:    "/swipe/middle/",
			body: models.SwipeRequest{
				Swiper:  "1234",
				Swipee:  "5678",
				Comment: "asdf"},
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: errorJson("not left or right"),
		},
		{
			name:   "non-numeric swiper",
			method: "POST",
			url:    "/swipe/left/",
			body: models.SwipeRequest{
				Swiper:  "abc1234",
				Swipee:  "5678",
				Comment: "asdf"},
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: errorJson("invalid swiper: abc1234"),
		},
		{
			name:   "non-numeric swipee",
			method: "POST",
			url:    "/swipe/left/",
			body: models.SwipeRequest{
				Swiper:  "1234",
				Swipee:  "abc5678",
				Comment: "asdf"},
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: errorJson("invalid swipee: abc5678"),
		},
		{
			name:   "comment too long",
			method: "POST",
			url:    "/swipe/right/",
			body: models.SwipeRequest{
				Swiper:  "1234",
				Swipee:  "5678",
				Comment: strings.Repeat("a", 257)}, // 257 bytes long
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: errorJson("comment too long"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tc.body)
			bodyReader := bytes.NewReader(bodyBytes)
			req, _ := http.NewRequest(tc.method, tc.url, bodyReader)

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Serve the request
			s.Handler.ServeHTTP(rr, req)

			resp := rr.Result()

			body, _ := io.ReadAll(resp.Body)
			message := string(body)

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			assert.Equal(t, tc.expectedMessage, message)
		})
	}
}

func TestGetUserStatsHandler(t *testing.T) {
	mockMetrics := mockMetrics.NewMetrics(t)
	mockPublisher := mockPublisher.NewPublisher(t)

	wantItem := &models.DynamoUserStats{
		UserId:      1234,
		NumLikes:    11,
		NumDislikes: 22,
		MatchList:   []int{1001, 1002, 1003},
	}
	mockDynamoClient := customMockDynamoClient(t, wantItem)

	databaseStub := &store.DatabaseClient{
		Client: mockDynamoClient,
		Table:  "SwipeData",
	}

	s := NewServer(":8080", mockMetrics, mockPublisher, databaseStub)
	req, _ := http.NewRequest("GET", "/stats/1234/", nil)
	rr := httptest.NewRecorder()
	s.Handler.ServeHTTP(rr, req)
	resp := rr.Result()

	assert.Equal(t, 200, resp.StatusCode)

	var stat models.UserStats
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &stat)
	assert.Equal(t, 11, stat.NumLikes)
	assert.Equal(t, 22, stat.NumDislikes)
}

// Convert a message to an error json with a newline
func errorJson(message string) string {
	encoded, _ := json.Marshal(
		&models.ErrorResponse{
			Message: message,
		})
	return fmt.Sprintf("%s\n", encoded)
}

func customMockDynamoClient(t *testing.T, wantItem *models.DynamoUserStats) *mockDynamo.DynamoClienter {
	mockDynamoClient := mockDynamo.NewDynamoClienter(t)

	av, _ := attributevalue.MarshalMap(wantItem)

	output := &dynamodb.GetItemOutput{
		Item: av,
	}

	mockDynamoClient.EXPECT().GetItem(
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.AnythingOfType("*dynamodb.GetItemInput"),
		mock.AnythingOfType("string")).Return(output, nil)

	return mockDynamoClient

}
