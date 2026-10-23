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
	statusCounts := make(map[int]int, len(r.StatusCounts))
	for status, count := range r.StatusCounts {
		statusCounts[status] = count
	}
	return RequestReport{
		RequestID:    r.RequestID,
		Name:         r.Name,
		Count:        r.Requests,
		AvgMs:        latencyMs(r.Histogram.Mean()),
		P90Ms:        latencyMs(r.Histogram.Percentile(90)),
		P95Ms:        latencyMs(r.Histogram.Percentile(95)),
		P99Ms:        latencyMs(r.Histogram.Percentile(99)),
		MinMs:        latencyMs(r.Histogram.Min),
		MaxMs:        latencyMs(r.Histogram.Max),
		Throughput:   r.Throughput(elapsed),
		ErrorRate:    r.ErrorRate(),
		StatusCounts: statusCounts,
	}
}

func latencyMs(latency time.Duration) float64 {
	return float64(latency.Microseconds()) / 1000
}
