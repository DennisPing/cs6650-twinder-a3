package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/DennisPing/cs6650-twinder-a3/httpclient/client"
	"github.com/DennisPing/cs6650-twinder-a3/httpclient/datagen"
	"github.com/DennisPing/cs6650-twinder-a3/lib/logger"
	"github.com/montanaflynn/stats"
)

const (
	maxWorkers  = 10
	numRequests = 1000
)

func main() {
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		logger.Fatal().Msg("SERVER_URL env variable not set")
	}

	port := os.Getenv("PORT") // Set the PORT to 8081 for local testing
	if port == "" {
		port = "8081" // Running in the cloud
	}

	// Health check endpoint
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		addr := fmt.Sprintf(":%s", port)
		logger.Fatal().Msg(http.ListenAndServe(addr, nil).Error())
	}()

	// Populate the task queue with tasks (token)
	taskQueue := make(chan struct{}, numRequests)
	for i := 0; i < numRequests; i++ {
		taskQueue <- struct{}{}
	}
	close(taskQueue) // Close the queue. Nothing is ever being put into the queue.

	var wg sync.WaitGroup

	responseTimes := make([][]time.Duration, maxWorkers)
	fetchResponseTimes := make([]time.Duration, 0)

	// Create a shared transport (1 TCP conn) since we have multiple clients for only 1 host
	sharedTransport := &http.Transport{
		MaxIdleConns:        maxWorkers + 1,
		MaxIdleConnsPerHost: maxWorkers + 1,
		IdleConnTimeout:     60 * time.Second,
	}

	// Spawn the worker pool
	workerPool := make([]*client.ApiClient, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		workerPool[i] = client.NewApiClient(sharedTransport, serverURL)
	}

	// Init the fetch client with default transport
	fetchClient := client.NewApiClient(sharedTransport, serverURL)

	logger.Info().Msgf("Using %d goroutines", maxWorkers)
	logger.Info().Msgf("Starting %d requests...", numRequests)
	startTime := time.Now()

	// Activate workers
	for i := 0; i < len(workerPool); i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			apiClient := workerPool[workerId]

			// Do tasks until taskQueue is empty. Then all workers can go home.
			for range taskQueue {
				direction := datagen.RandDirection(apiClient.Rng)
				t0 := time.Now()
				apiClient.SwipeLeftOrRight(direction) // The actual HTTP request
				time.Sleep(500 * time.Millisecond)
				t1 := time.Since(t0)
				responseTimes[workerId] = append(responseTimes[workerId], t1) // Thread safe
			}
		}(i)
	}

	// Activate the fetch client
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop() // ticker should stop at ctx.Done() but use "defer" for edge cases

		toggle := true
		for {
			select {
			case <-ctx.Done(): // Quit
				return
			case <-ticker.C:
				t0 := time.Now()
				if toggle {
					fetchClient.GetUserStats()
				} else {
					fetchClient.GetMatches()
				}
				t1 := time.Since(t0)
				fetchResponseTimes = append(fetchResponseTimes, t1)
				toggle = !toggle
			}
		}
	}()

	wg.Wait() // Wait for all workers to finish tasks

	cancel() // Cancel the fetch client goroutine

	duration := time.Since(startTime)

	// Calculate metrics
	var postSuccessCount uint64
	var postErrorCount uint64
	for _, worker := range workerPool {
		postSuccessCount += worker.PostSuccessCount
		postErrorCount += worker.PostErrorCount
	}
	throughput := float64(postSuccessCount) / duration.Seconds()

	fmt.Println("Done!")
	logger.Info().Msgf("Total run time: %v", duration)
	logger.Info().Msgf("Throughput: %.2f req/sec", throughput)

	allResponseTimes := make([]float64, 0, numRequests)
	for _, slice := range responseTimes { // Convert all time.Duration to float64
		for _, rt := range slice {
			rtFloat := float64(rt.Milliseconds())
			allResponseTimes = append(allResponseTimes, rtFloat)
		}
	}

	mean, _ := stats.Mean(allResponseTimes)
	median, _ := stats.Median(allResponseTimes)
	p99, _ := stats.Percentile(allResponseTimes, 99)
	min, _ := stats.Min(allResponseTimes)
	max, _ := stats.Max(allResponseTimes)

	fmt.Println("POST request client metrics")
	logger.Info().Msgf("POST success count: %d", postSuccessCount)
	logger.Info().Msgf("POST error count: %d", postErrorCount)
	logger.Info().Msgf("Mean response time: %.2f ms", mean)
	logger.Info().Msgf("Median response time: %.2f ms", median)
	logger.Info().Msgf("P99 response time: %.2f ms", p99)
	logger.Info().Msgf("Min response time: %.2f ms", min)
	logger.Info().Msgf("Max response time: %.2f ms", max)

	allFetchResponseTimes := make([]float64, 0, len(fetchResponseTimes))
	for _, rt := range fetchResponseTimes { // Convert all time.Duration to float64
		rtFloat := float64(rt.Milliseconds())
		allFetchResponseTimes = append(allFetchResponseTimes, rtFloat)
	}

	mean, _ = stats.Mean(allFetchResponseTimes)
	median, _ = stats.Median(allFetchResponseTimes)
	p99, _ = stats.Percentile(allFetchResponseTimes, 99)
	min, _ = stats.Min(allFetchResponseTimes)
	max, _ = stats.Max(allFetchResponseTimes)

	fmt.Println("GET request client metrics")
	logger.Info().Msgf("GET success count: %d", fetchClient.GetSuccessCount)
	logger.Info().Msgf("GET success count: %d", fetchClient.GetErrorCount)
	logger.Info().Msgf("Mean response time: %.2f ms", mean)
	logger.Info().Msgf("Median response time: %.2f ms", median)
	logger.Info().Msgf("P99 response time: %.2f ms", p99)
	logger.Info().Msgf("Min response time: %.2f ms", min)
	logger.Info().Msgf("Max response time: %.2f ms", max)
}
