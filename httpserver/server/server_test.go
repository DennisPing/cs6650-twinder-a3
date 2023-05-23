package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mockMetrics "github.com/DennisPing/cs6650-twinder-a3/httpserver/metrics/mocks"
	mockPublisher "github.com/DennisPing/cs6650-twinder-a3/httpserver/rmqproducer/mocks"
	"github.com/DennisPing/cs6650-twinder-a3/lib/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandlers(t *testing.T) {
	// Mock the internal Metrics
	mockMetrics := mockMetrics.NewMetrics(t)
	mockMetrics.On("IncrementThroughput").Return()

	// Mock the internal Publisher
	mockPublisher := mockPublisher.NewPublisher(t)
	mockPublisher.On("Publish",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]string"),
		mock.AnythingOfType("func(*rabbitmq.PublishOptions)"),
		mock.AnythingOfType("func(*rabbitmq.PublishOptions)")).Return(nil)

	s := NewServer(":8080", mockMetrics, mockPublisher)

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
			body: models.SwipePayload{
				Swiper:  "1234",
				Swipee:  "5678",
				Comment: "asdf"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "swipe right",
			method: "POST",
			url:    "/swipe/right/",
			body: models.SwipePayload{
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

func TestHandlersError(t *testing.T) {
	// Mock the internal Metrics
	mockMetrics := mockMetrics.NewMetrics(t)

	// Mock the internal Publisher
	mockPublisher := mockPublisher.NewPublisher(t)

	s := NewServer(":8080", mockMetrics, mockPublisher)

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
			body: models.SwipePayload{
				Swiper:  "1234",
				Swipee:  "5678",
				Comment: "asdf"},
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "not left or right\n",
		},
		{
			name:   "non-numeric swiper",
			method: "POST",
			url:    "/swipe/left/",
			body: models.SwipePayload{
				Swiper:  "abc1234",
				Swipee:  "5678",
				Comment: "asdf"},
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "invalid swiper: abc1234\n",
		},
		{
			name:   "non-numeric swipee",
			method: "POST",
			url:    "/swipe/left/",
			body: models.SwipePayload{
				Swiper:  "1234",
				Swipee:  "abc5678",
				Comment: "asdf"},
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "invalid swipee: abc5678\n",
		},
		{
			name:   "comment too long",
			method: "POST",
			url:    "/swipe/right/",
			body: models.SwipePayload{
				Swiper:  "1234",
				Swipee:  "5678",
				Comment: strings.Repeat("a", 257)}, // 257 bytes long
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "comment too long\n",
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

			respBody, _ := io.ReadAll(resp.Body)
			message := string(respBody)

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			assert.Equal(t, tc.expectedMessage, message)
		})
	}
}
