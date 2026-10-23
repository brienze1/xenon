package entity

import "time"

// RequestReport is the figures of a finished or running performance run, overall or for one request.
type RequestReport struct {
	StatusCounts map[int]int
	RequestID    string
	Name         string
	Count        int
	AvgMs        float64
	P90Ms        float64
	P95Ms        float64
	P99Ms        float64
	MinMs        float64
	MaxMs        float64
	Throughput   float64
	ErrorRate    float64
}

func NewRequestReport(stats *RequestStats, elapsed time.Duration) RequestReport {
	statusCounts := make(map[int]int, len(stats.StatusCounts))
	for status, count := range stats.StatusCounts {
		statusCounts[status] = count
	}
	return RequestReport{
		RequestID:    stats.RequestID,
		Name:         stats.Name,
		Count:        stats.Requests,
		AvgMs:        latencyMs(stats.Histogram.Mean()),
		P90Ms:        latencyMs(stats.Histogram.Percentile(90)),
		P95Ms:        latencyMs(stats.Histogram.Percentile(95)),
		P99Ms:        latencyMs(stats.Histogram.Percentile(99)),
		MinMs:        latencyMs(stats.Histogram.Min),
		MaxMs:        latencyMs(stats.Histogram.Max),
		Throughput:   stats.Throughput(elapsed),
		ErrorRate:    stats.ErrorRate(),
		StatusCounts: statusCounts,
	}
}
