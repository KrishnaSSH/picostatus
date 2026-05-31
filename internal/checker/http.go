package checker

import (
	"context"
	"net/http"
	"time"
)

type HTTPChecker struct {
	URL     string
	Timeout time.Duration
}

func (h HTTPChecker) Run(ctx context.Context) Result {
	checkCtx, cancel := context.WithTimeout(ctx, h.Timeout)
	defer cancel()

	start := time.Now()

	req, err := http.NewRequestWithContext(checkCtx, http.MethodGet, h.URL, nil)
	if err != nil {
		return Result{Status: "down", Error: err.Error()}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Result{Status: "down", LatencyMS: time.Since(start).Milliseconds(), Error: err.Error()}
	}
	defer resp.Body.Close()

	status := "down"
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		status = "up"
	}

	return Result{
		Status:    status,
		LatencyMS: time.Since(start).Milliseconds(),
	}
}
