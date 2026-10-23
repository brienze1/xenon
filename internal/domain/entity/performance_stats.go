package entity

import (
	"sync"
	"time"
)

type secondTally struct {
	latency     time.Duration
	activeUsers int
	requests    int
	responses   int
	errors      int
}

// PerformanceStats aggregates a performance run; it is safe for concurrent use by the virtual users.
type PerformanceStats struct {
	PerRequest []RequestStats
	seconds    []secondTally
	Overall    RequestStats
	mutex      sync.Mutex
}

func NewPerformanceStats(rows []RunPlanRow) *PerformanceStats {
	perRequest := make([]RequestStats, 0, len(rows))
	for _, row := range rows {
		perRequest = append(perRequest, NewRequestStats(row.RequestID, row.Name))
	}
	return &PerformanceStats{Overall: NewRequestStats("", ""), PerRequest: perRequest, seconds: []secondTally{}}
}

func (p *PerformanceStats) Record(sample PerformanceSample) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.Overall.Record(sample)
	if sample.Row >= 0 && sample.Row < len(p.PerRequest) {
		p.PerRequest[sample.Row].Record(sample)
	}
	tally := p.second(sample.Second)
	tally.requests++
	if sample.IsError() {
		tally.errors++
	}
	if sample.Responded() {
		tally.responses++
		tally.latency += sample.Latency
	}
}

// Observe keeps the highest active user count seen during a second.
func (p *PerformanceStats) Observe(second, activeUsers int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	tally := p.second(second)
	tally.activeUsers = max(tally.activeUsers, activeUsers)
}

func (p *PerformanceStats) Metrics(activeUsers int, elapsed time.Duration) PerformanceMetrics {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return PerformanceMetrics{
		ActiveUsers: activeUsers,
		Requests:    p.Overall.Requests,
		AvgMs:       latencyMs(p.Overall.Histogram.Mean()),
		Throughput:  p.Overall.Throughput(elapsed),
		ErrorRate:   p.Overall.ErrorRate(),
	}
}

// Report gives the figures over the elapsed time, with exactly one series entry for each of the seconds.
func (p *PerformanceStats) Report(elapsed time.Duration, seconds int) PerformanceReport {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	perRequest := make([]RequestReport, 0, len(p.PerRequest))
	for index := range p.PerRequest {
		perRequest = append(perRequest, p.PerRequest[index].Report(elapsed))
	}
	series := make([]PerformanceSecond, 0, seconds)
	for second := range seconds {
		series = append(series, p.secondReport(second))
	}
	return PerformanceReport{Overall: p.Overall.Report(elapsed), PerRequest: perRequest, Series: series}
}

func (p *PerformanceStats) secondReport(second int) PerformanceSecond {
	report := PerformanceSecond{Second: second}
	if second >= len(p.seconds) {
		return report
	}
	tally := p.seconds[second]
	report.ActiveUsers = tally.activeUsers
	report.Requests = tally.requests
	report.Errors = tally.errors
	if tally.responses > 0 {
		report.AvgMs = latencyMs(tally.latency / time.Duration(tally.responses))
	}
	return report
}

func (p *PerformanceStats) second(second int) *secondTally {
	for len(p.seconds) <= second {
		p.seconds = append(p.seconds, secondTally{})
	}
	return &p.seconds[second]
}
