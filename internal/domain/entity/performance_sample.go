package entity

import "time"

type PerformanceSample struct {
	Latency time.Duration
	Row     int
	Second  int
	Status  int
}

func NewPerformanceSample(result RunRowResult, latency time.Duration, row, second int) PerformanceSample {
	return PerformanceSample{Latency: latency, Row: row, Second: second, Status: result.Status}
}

// IsError is true for a request that could not be sent or got a status of 400 or more.
func (p PerformanceSample) IsError() bool {
	return p.Status == 0 || p.Status >= 400
}

func (p PerformanceSample) Responded() bool {
	return p.Status > 0
}
