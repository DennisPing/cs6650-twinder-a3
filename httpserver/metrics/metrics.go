package metrics

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/ingest"
)

//go:generate mockery --name=Metrics --filename=mock_metrics.go
type Metrics interface {
	IncrementThroughput()
	GetThroughput() uint64
	SendMetrics() error
}

type AxiomMetrics struct {
	client      *axiom.Client
	ServerId    string
	DatasetName string
	Throughput  uint64
	Mutex       sync.Mutex
}

// Create a new AxiomMetrics client which implements the Metrics interface
func NewMetricsClient() (*AxiomMetrics, error) {
	serverId := os.Getenv("RAILWAY_REPLICA_ID")
	apiToken := os.Getenv("AXIOM_API_TOKEN")
	datasetName := os.Getenv("AXIOM_DATASET")

	if apiToken == "" || datasetName == "" {
		return nil, errors.New("you forgot to set the AXIOM env variables")
	}
	client, err := axiom.NewClient(
		axiom.SetAPITokenConfig(apiToken),
	)
	if err != nil {
		return nil, err
	}
	return &AxiomMetrics{
		client:      client,
		ServerId:    serverId,
		DatasetName: datasetName,
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
	ctx := context.Background()

	if _, err := m.client.IngestEvents(ctx, m.DatasetName, []axiom.Event{
		{ingest.TimestampField: time.Now(), "ServerId": m.ServerId, "Throughput": throughput},
	}); err != nil {
		return err
	}
	return nil
}
