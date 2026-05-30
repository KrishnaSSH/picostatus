package storage

import "time"

type Check struct {
	ID        int64
	Name      string
	Target    string
	Interval  time.Duration
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Result struct {
	ID        int64
	CheckID   int64
	Success   bool
	LatencyMS int64
	Error     string
	CheckedAt time.Time
}
