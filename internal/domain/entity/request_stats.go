package entity

import "time"

type RequestStats struct {
	StatusCounts map[int]int
	RequestID    string
	Name         string
	Histogram    LatencyHistogram
	Requests     int
	Errors       int
}

func NewRequestStats(requestId, name string) RequestStats {
	return RequestStats{RequestID: requestId, Name: name, Histogram: NewLatencyHistogram(), StatusCounts: map[int]int{}}
}

func (r *RequestStats) Record(sample PerformanceSample) {
	r.Requests++
	if sample.IsError() {
		r.Errors++
	}
	if sample.Responded() {
		r.StatusCounts[sample.Status]++
		r.Histogram.Record(sample.Latency)
	}
}

// Throughput is the completed requests per second of elapsed run time.
func (r *RequestStats) Throughput(elapsed time.Duration) float64 {
	if elapsed <= 0 {
		return 0
	}
	return float64(r.Requests) / elapsed.Seconds()
}

func (r *RequestStats) ErrorRate() float64 {
	if r.Requests == 0 {
		return 0
	}
	return float64(r.Errors) / float64(r.Requests)
}

func (r *RequestStats) Report(elapsed time.Duration) RequestReport {
	return NewRequestReport(r, elapsed)
}

func latencyMs(latency time.Duration) float64 {
	return float64(latency.Microseconds()) / 1000
}
