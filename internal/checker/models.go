package checker

type Result struct {
	Success   bool
	LatencyMS int64
	Error     string
}
