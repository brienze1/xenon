package entity

// PerformanceMetrics is the live view of a performance run sent with its progress.
type PerformanceMetrics struct {
	ActiveUsers int
	Requests    int
	AvgMs       float64
	Throughput  float64
	ErrorRate   float64
}
