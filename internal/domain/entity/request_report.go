package entity

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
