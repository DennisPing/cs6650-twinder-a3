package metrics

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/DennisPing/cs6650-twinder-a3/lib/models"
)

//go:generate mockery --name=Metrics --filename=mock_metrics.go
type Metrics interface {
	IncrementThroughput()
	GetThroughput() uint64
	SendMetrics() error
}

type AxiomMetrics struct {
	ServerId    string
	DatasetName string
	ApiToken    string
	IngestUrl   string
	Throughput  uint64
	Mutex       sync.Mutex
}

func NewMetrics() (*AxiomMetrics, error) {
	serverId := os.Getenv("RAILWAY_REPLICA_ID")
	datasetName := os.Getenv("AXIOM_DATASET")
	apiToken := os.Getenv("AXIOM_API_TOKEN")
	ingestUrl := "https://api.axiom.co/v1/datasets/%s/ingest"

	if datasetName == "" || apiToken == "" {
		return nil, errors.New("you forgot to set the AXIOM env variables")
	}
	return &AxiomMetrics{
		ServerId:    serverId,
		DatasetName: datasetName,
		ApiToken:    apiToken,
		IngestUrl:   ingestUrl,
	}, nil
}

// Increment the throughput count
func (m *AxiomMetrics) IncrementThroughput() {
	m.Mutex.Lock()
	m.Throughput++
	m.Mutex.Unlock()
}

// Return the throughput and reset the count
func (m *AxiomMetrics) GetThroughput() uint64 {
	m.Mutex.Lock()
	throughput := m.Throughput
	m.Throughput = 0
	m.Mutex.Unlock()
	return throughput
}

// Send the metrics over to Axiom
func (m *AxiomMetrics) SendMetrics() error {
	throughput := m.GetThroughput()
	payload := models.AxiomPayload{
		Time:       time.Now().Format(time.RFC3339Nano),
		ServerId:   m.ServerId,
		Throughput: throughput,
	}

	jsonPayload, err := json.Marshal([]models.AxiomPayload{payload})
	if err != nil {
		return err
	}

	url := fmt.Sprintf(m.IngestUrl, m.DatasetName)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+m.ApiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("%s", resp.Status)
	}

	return nil
}
