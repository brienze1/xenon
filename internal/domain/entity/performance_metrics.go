package entity

import "time"

type PerformanceMetrics struct {
	ActiveUsers int
	Requests    int
	AvgMs       float64
	Throughput  float64
	ErrorRate   float64
}

func NewPerformanceMetrics(activeUsers int, overall *RequestStats, elapsed time.Duration) PerformanceMetrics {
	return PerformanceMetrics{
		ActiveUsers: activeUsers,
		Requests:    overall.Requests,
		AvgMs:       latencyMs(overall.Histogram.Mean()),
		Throughput:  overall.Throughput(elapsed),
		ErrorRate:   overall.ErrorRate(),
	}
}
