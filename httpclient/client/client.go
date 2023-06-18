package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/DennisPing/cs6650-twinder-a3/httpclient/datagen"
	"github.com/DennisPing/cs6650-twinder-a3/lib/logger"
	"github.com/DennisPing/cs6650-twinder-a3/lib/models"
)

var zlog = logger.GetLogger()

// An api client that has a random number generator
type ApiClient struct {
	ServerUrl    string
	HttpClient   *http.Client
	Rng          *rand.Rand
	SuccessCount uint64
	ErrorCount   uint64
}

func NewApiClient(transport *http.Transport, serverUrl string) *ApiClient {
	return &ApiClient{
		ServerUrl: serverUrl,
		HttpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		},
		Rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// POST /swipe/{leftorright}/
func (client *ApiClient) SwipeLeftOrRight(direction string) {
	swipeRequest := models.SwipeRequest{
		Swiper:  strconv.Itoa(datagen.RandInt(client.Rng, 1, 5000)),
		Swipee:  strconv.Itoa(datagen.RandInt(client.Rng, 1, 1_000_000)),
		Comment: datagen.RandComment(client.Rng, 256),
	}
	endpoint := fmt.Sprintf("%s/swipe/%s/", client.ServerUrl, direction)

	req := client.newPostRequest(endpoint, swipeRequest)
	resp, err := client.sendRequest(req, 5)
	if err != nil {
		client.ErrorCount += 1
		zlog.Error().Err(err).Str("method", "POST").Str("endpoint", endpoint).Msg("max retries hit")
		return
	}
	defer resp.Body.Close()

	// StatusCode should be 200 or 201, else log warn
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		client.SuccessCount += 1
		zlog.Debug().Str("method", "POST").Str("endpoint", endpoint).Int("code", resp.StatusCode).Msg("response")
	} else {
		client.ErrorCount += 1
		zlog.Warn().Str("method", "POST").Str("endpoint", endpoint).Int("code", resp.StatusCode).Msg("response")
	}
}

// GET /stats/{userId}/
func (client *ApiClient) GetUserStats() {
	userId := datagen.RandInt(client.Rng, 1, 5000)
	endpoint := fmt.Sprintf("%s/stats/%d/", client.ServerUrl, userId)

	req := client.newGetRequest(endpoint)
	resp, err := client.sendRequest(req, 5)
	if err != nil {
		client.ErrorCount += 1
		zlog.Error().Err(err).Str("method", "GET").Str("endpoint", endpoint).Msg("max retries hit")
		return
	}
	defer resp.Body.Close()

	// StatusCode should be 200 or 404, else log warn
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		client.SuccessCount += 1
		zlog.Debug().Str("method", "GET").Str("endpoint", endpoint).Int("code", resp.StatusCode).Msg("response")
	} else {
		client.ErrorCount += 1
		zlog.Warn().Str("method", "GET").Str("endpoint", endpoint).Int("code", resp.StatusCode).Msg("response")
	}
}

// GET /matches/{userId}/
func (client *ApiClient) GetMatches() {
	userId := datagen.RandInt(client.Rng, 1, 5000)
	endpoint := fmt.Sprintf("%s/matches/%d/", client.ServerUrl, userId)

	req := client.newGetRequest(endpoint)
	resp, err := client.sendRequest(req, 5)
	if err != nil {
		client.ErrorCount += 1
		zlog.Error().Err(err).Str("method", "GET").Str("endpoint", endpoint).Msg("max retries hit")
		return
	}
	defer resp.Body.Close()

	// StatusCode should be 200 or 404, else log warn
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		client.SuccessCount += 1
		zlog.Debug().Str("method", "GET").Str("endpoint", endpoint).Int("code", resp.StatusCode).Msg("response")
	} else {
		client.ErrorCount += 1
		zlog.Warn().Str("method", "GET").Str("endpoint", endpoint).Int("code", resp.StatusCode).Msg("response")
	}
}

// Create a new HTTP GET request
func (client *ApiClient) newGetRequest(url string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to build GET request") // programmer error
	}
	return req
}

// Create a new HTTP POST request
func (client *ApiClient) newPostRequest(url string, data interface{}) *http.Request {
	body, err := json.Marshal(data)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to build marshal data") // programmer error
	}
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to build POST request") // programmer error
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))

	return req
}

// Send HTTP request with retry limit
func (client *ApiClient) sendRequest(req *http.Request, maxRetries int) (*http.Response, error) {
	baseBackoff := 100 * time.Millisecond

	var resp *http.Response
	var err error
	for i := 1; i <= maxRetries; i++ {
		resp, err = client.HttpClient.Do(req)
		if err == nil {
			break // Successful API call
		}
		// Exponential backoff with jitter
		backoffDuration := time.Duration(math.Pow(2, float64(i))) * baseBackoff
		sleepDuration := backoffDuration + time.Duration(client.Rng.Int63n(1000))*time.Millisecond
		time.Sleep(sleepDuration)
	}
	return resp, err
}
