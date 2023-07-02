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
	maxWorkers  = 50
	numRequests = 100_000
)

var zlog = logger.GetLogger()

func main() {
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		zlog.Fatal().Msg("SERVER_URL env variable not set")
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
		if err := http.ListenAndServe(addr, nil); err != nil {
			zlog.Fatal().Err(err).Msg("health check crashed")
		}
	}()

	// Populate the task queue with tasks (token)
	taskQueue := make(chan struct{}, numRequests)
	for i := 0; i < numRequests; i++ {
		taskQueue <- struct{}{}
	}
	close(taskQueue) // Close the queue. Nothing is ever being put into the queue.

	responseTimes := make([][]time.Duration, maxWorkers)
	fetchResponseTimes := make([][]time.Duration, 5)

	// Create a shared transport (1 TCP conn) since we have multiple clients for only 1 host
	sharedTransport := &http.Transport{
		MaxIdleConns:        maxWorkers,
		MaxIdleConnsPerHost: maxWorkers,
		IdleConnTimeout:     60 * time.Second,
	}

	// Spawn the worker pool
	workerPool := make([]*client.ApiClient, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		workerPool[i] = client.NewApiClient(sharedTransport, serverURL)
	}

	// Crate another shared transport for the fetch clients
	fetchTransport := &http.Transport{
		MaxIdleConns:        5,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     60 * time.Second,
	}

	// Spawn the fetch pool
	fetchPool := make([]*client.ApiClient, 5)
	for i := 0; i < 5; i++ {
		fetchPool[i] = client.NewApiClient(fetchTransport, serverURL)
	}

	// Start the main actions
	zlog.Info().Msgf("Using %d goroutines", maxWorkers)
	zlog.Info().Msgf("Starting %d requests...", numRequests)
	startTime := time.Now()

	var wg sync.WaitGroup

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
				t1 := time.Since(t0)
				responseTimes[workerId] = append(responseTimes[workerId], t1) // Thread safe
			}
		}(i)
	}

	// Activate the fetch clients
	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < len(fetchPool); i++ {
		go func(id int) {
			fetchClient := fetchPool[id]
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop() // ticker should stop at ctx.Done() but use "defer" for edge cases
			toggle := true

			// Loop forever until the ctx is canceled
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
					fetchResponseTimes[id] = append(fetchResponseTimes[id], t1)
					toggle = !toggle
				}
			}
		}(i)
	}

	wg.Wait() // Wait for all workers to finish tasks

	cancel() // Cancel the fetch client goroutine

	duration := time.Since(startTime)

	// Calculate metrics for worker clients
	var postSuccessCount uint64
	var postErrorCount uint64
	for _, worker := range workerPool {
		postSuccessCount += worker.SuccessCount
		postErrorCount += worker.ErrorCount
	}
	throughput := float64(postSuccessCount) / duration.Seconds()

	fmt.Println("Done!")
	zlog.Info().Msgf("Total run time: %v", duration)
	zlog.Info().Msgf("Throughput: %.2f req/sec", throughput)

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
	zlog.Info().Msgf("POST success count: %d", postSuccessCount)
	zlog.Info().Msgf("POST error count: %d", postErrorCount)
	zlog.Info().Msgf("Mean response time: %.2f ms", mean)
	zlog.Info().Msgf("Median response time: %.2f ms", median)
	zlog.Info().Msgf("P99 response time: %.2f ms", p99)
	zlog.Info().Msgf("Min response time: %.2f ms", min)
	zlog.Info().Msgf("Max response time: %.2f ms", max)

	// Calculate metrics for fetch clients
	var getSuccessCount uint64
	var getErrorCount uint64
	for _, fetcher := range fetchPool {
		getSuccessCount += fetcher.SuccessCount
		getErrorCount += fetcher.ErrorCount
	}

	allFetchResponseTimes := make([]float64, 0)
	for _, slice := range fetchResponseTimes { // Convert all time.Duration to float64
		for _, rt := range slice {
			rtFloat := float64(rt.Milliseconds())
			allFetchResponseTimes = append(allFetchResponseTimes, rtFloat)
		}
	}

	mean, _ = stats.Mean(allFetchResponseTimes)
	median, _ = stats.Median(allFetchResponseTimes)
	p99, _ = stats.Percentile(allFetchResponseTimes, 99)
	min, _ = stats.Min(allFetchResponseTimes)
	max, _ = stats.Max(allFetchResponseTimes)

	fmt.Println("GET request client metrics")
	zlog.Info().Msgf("GET success count: %d", getSuccessCount)
	zlog.Info().Msgf("GET error count: %d", getErrorCount)
	zlog.Info().Msgf("Mean response time: %.2f ms", mean)
	zlog.Info().Msgf("Median response time: %.2f ms", median)
	zlog.Info().Msgf("P99 response time: %.2f ms", p99)
	zlog.Info().Msgf("Min response time: %.2f ms", min)
	zlog.Info().Msgf("Max response time: %.2f ms", max)
}
