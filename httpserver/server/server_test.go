package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	_mockRepo "github.com/DennisPing/cs6650-twinder-a3/httpserver/db/dynamodb_repo"
	_mockService "github.com/DennisPing/cs6650-twinder-a3/httpserver/db/dynamodb_repo/mocks"
	_mockMetrics "github.com/DennisPing/cs6650-twinder-a3/httpserver/metrics/mocks"
	_mockPublisher "github.com/DennisPing/cs6650-twinder-a3/httpserver/rmqproducer/mocks"
	"github.com/DennisPing/cs6650-twinder-a3/lib/models"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPostSwipe(t *testing.T) {
	// Create a quiet logger
	quietLogger := zerolog.New(io.Discard)

	// Mock the internal Metrics
	mockMetrics := _mockMetrics.NewMetrics(t)
	mockMetrics.EXPECT().IncrementThroughput().Return()

	// Mock the internal Publisher
	mockPublisher := _mockPublisher.NewPublisher(t)
	mockPublisher.EXPECT().Publish(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Mock the remote DynamoDB service
	mockService := _mockService.NewDynamoDBServicer(t)

	// Real repository with a mock service inside
	repo := _mockRepo.NewDynamoDBRepo(mockService)

	s := NewServer(":8080", quietLogger, mockMetrics, mockPublisher, repo)

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
			url:    "/swipe/left",
			body: models.SwipeRequest{
				Swiper:  "1234",
				Swipee:  "5678",
				Comment: "asdf"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "swipe right",
			method: "POST",
			url:    "/swipe/right",
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

			rr := httptest.NewRecorder()

			s.Handler.ServeHTTP(rr, req)

			resp := rr.Result()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestPostSwipeError(t *testing.T) {
	// Create a quiet logger
	quietLogger := zerolog.New(io.Discard)

	// Mock the internal Metrics
	mockMetrics := _mockMetrics.NewMetrics(t)

	// Mock the internal Publisher
	mockPublisher := _mockPublisher.NewPublisher(t)

	// Mock the remote DynamoDB service
	mockService := _mockService.NewDynamoDBServicer(t)

	// Real repository with a mock service inside
	repo := _mockRepo.NewDynamoDBRepo(mockService)

	s := NewServer(":8080", quietLogger, mockMetrics, mockPublisher, repo)

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
			url:    "/swipe/middle",
			body: models.SwipeRequest{
				Swiper:  "1234",
				Swipee:  "5678",
				Comment: "asdf"},
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: errorJson("not left or right: middle"),
		},
		{
			name:   "non-numeric swiper",
			method: "POST",
			url:    "/swipe/left",
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
			url:    "/swipe/left",
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
			url:    "/swipe/right",
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

			rr := httptest.NewRecorder()

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
	quietLogger := zerolog.New(io.Discard)
	mockMetrics := _mockMetrics.NewMetrics(t)
	mockPublisher := _mockPublisher.NewPublisher(t)

	tests := []struct {
		name             string
		method           string
		url              string
		itemOutput       *dynamodb.GetItemOutput
		itemOutputError  error
		found            bool
		expectedStatus   int
		expectedLikes    int
		expectedDislikes int
	}{
		{
			name:   "Get user stats happy path",
			method: "GET",
			url:    "/stats/1234",
			itemOutput: &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"UserId":      &types.AttributeValueMemberN{Value: strconv.Itoa(1234)},
					"NumLikes":    &types.AttributeValueMemberN{Value: strconv.Itoa(11)},
					"NumDislikes": &types.AttributeValueMemberN{Value: strconv.Itoa(22)},
					"MatchList":   &types.AttributeValueMemberNS{Value: []string{"1001", "1002", "1003"}},
				},
			},
			itemOutputError:  nil,
			found:            true,
			expectedStatus:   200,
			expectedLikes:    11,
			expectedDislikes: 22,
		},
		{
			name:             "User id not found",
			method:           "GET",
			url:              "/stats/1234",
			itemOutput:       &dynamodb.GetItemOutput{Item: nil},
			itemOutputError:  nil,
			found:            false,
			expectedStatus:   404,
			expectedLikes:    0,
			expectedDislikes: 0,
		},
		{
			name:             "AWS internal service error",
			method:           "GET",
			url:              "/stats/1234",
			itemOutput:       &dynamodb.GetItemOutput{Item: nil},
			itemOutputError:  errors.New("AWS broke"),
			found:            false,
			expectedStatus:   500,
			expectedLikes:    0,
			expectedDislikes: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := _mockService.NewDynamoDBServicer(t)
			mockService.EXPECT().GetItem(mock.Anything, mock.Anything, mock.Anything).Return(tc.itemOutput, tc.itemOutputError)

			// Real repository with a mock service inside
			repo := _mockRepo.NewDynamoDBRepo(mockService)

			s := NewServer(":8080", quietLogger, mockMetrics, mockPublisher, repo)

			bodyReader := bytes.NewReader(nil)
			req, _ := http.NewRequest(tc.method, tc.url, bodyReader)

			rr := httptest.NewRecorder()

			s.Handler.ServeHTTP(rr, req)

			resp := rr.Result()
			body, _ := io.ReadAll(resp.Body)
			var stats models.UserStats
			_ = json.Unmarshal(body, &stats)

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			assert.Equal(t, tc.expectedLikes, stats.NumLikes)
			assert.Equal(t, tc.expectedDislikes, stats.NumDislikes)
		})
	}
}

// Convert a message to an error json
func errorJson(message string) string {
	encoded, _ := json.Marshal(
		&models.ErrorResponse{
			Message: message,
		})
	return string(encoded)
}
