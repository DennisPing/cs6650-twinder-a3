package models

type AxiomPayload struct {
	Time       string `json:"_time"`
	ServerId   string `json:"serverId"`
	Throughput uint64 `json:"throughput"`
}
