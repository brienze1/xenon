package entity

import "time"

type secondTally struct {
	latency     time.Duration
	activeUsers int
	requests    int
	responses   int
	errors      int
}

func (s *secondTally) Report(second int) PerformanceSecond {
	return NewPerformanceSecond(second, s.activeUsers, s.requests, s.errors, s.responses, s.latency)
}
