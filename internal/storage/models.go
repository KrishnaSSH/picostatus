package storage

import "time"

type Status string

const (
	StatusUp      Status = "up"
	StatusDown    Status = "down"
	StatusUnknown Status = "unknown"
)

type Check struct {
	ID       int64
	Name     string
	Target   string
	Interval time.Duration
	Timeout  time.Duration
}

type Result struct {
	ID        int64
	CheckID   int64
	Status    Status
	LatencyMS int64
	Error     string
	CreatedAt time.Time
}

type Uptime struct {
	CheckID   int64
	Uptime1h  float64
	Uptime24h float64
	Uptime7d  float64
}
