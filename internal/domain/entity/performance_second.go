package entity

import "time"

type PerformanceSecond struct {
	Second      int
	ActiveUsers int
	Requests    int
	Errors      int
	AvgMs       float64
}

// NewPerformanceSecond averages the total latency over the responses received during the second.
func NewPerformanceSecond(second, activeUsers, requests, errors, responses int, latency time.Duration) PerformanceSecond {
	report := PerformanceSecond{Second: second, ActiveUsers: activeUsers, Requests: requests, Errors: errors}
	if responses > 0 {
		report.AvgMs = latencyMs(latency / time.Duration(responses))
	}
	return report
}
