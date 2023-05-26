package models

// The time frequency is controlled by the caller (eg. 1 sec, 5 sec, 30 sec)

// How many database writes per unit of time
type ConsumerWriteThroughput struct {
	Time       string `json:"_time"`
	ConsumerId string `json:"consumerId"`
	Throughput uint64 `json:"throughput"`
}

// How many server requests handled per unit of time
type ServerThroughput struct {
	Time       string `json:"_time"`
	ServerId   string `json:"serverId"`
	Throughput uint64 `json:"throughput"`
}
