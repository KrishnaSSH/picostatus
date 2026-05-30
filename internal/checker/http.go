package checker

import (
	"context"
	"net/http"
	"time"
)

type HTTPChecker struct {
	URL string
}

func (h HTTPChecker) Run(ctx context.Context) Result {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.URL, nil)
	if err != nil {
		return Result{
			Success: false,
			Error:   err.Error(),
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Result{
			Success: false,
			Error:   err.Error(),
		}
	}

	defer resp.Body.Close()

	return Result{
		Success:   resp.StatusCode >= 200 && resp.StatusCode < 300,
		LatencyMS: time.Since(start).Milliseconds(),
	}
}
