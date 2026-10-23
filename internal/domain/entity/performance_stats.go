package entity

import "time"

type PerformanceStats struct {
	PerRequest []RequestStats
	seconds    []secondTally
	Overall    RequestStats
}

func NewPerformanceStats(rows []RunPlanRow) *PerformanceStats {
	perRequest := make([]RequestStats, 0, len(rows))
	for _, row := range rows {
		perRequest = append(perRequest, NewRequestStats(row.RequestID, row.Name))
	}
	return &PerformanceStats{Overall: NewRequestStats("", ""), PerRequest: perRequest, seconds: []secondTally{}}
}

func (p *PerformanceStats) Record(sample PerformanceSample) {
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

func (p *PerformanceStats) Observe(second, activeUsers int) {
	tally := p.second(second)
	tally.activeUsers = max(tally.activeUsers, activeUsers)
}

func (p *PerformanceStats) Metrics(activeUsers int, elapsed time.Duration) PerformanceMetrics {
	return NewPerformanceMetrics(activeUsers, &p.Overall, elapsed)
}

func (p *PerformanceStats) Report(elapsed time.Duration, seconds int) PerformanceReport {
	perRequest := make([]RequestReport, 0, len(p.PerRequest))
	for index := range p.PerRequest {
		perRequest = append(perRequest, p.PerRequest[index].Report(elapsed))
	}
	series := make([]PerformanceSecond, 0, seconds)
	for second := range seconds {
		series = append(series, p.secondReport(second))
	}
	return NewPerformanceReport(p.Overall.Report(elapsed), perRequest, series)
}

func (p *PerformanceStats) secondReport(second int) PerformanceSecond {
	if second >= len(p.seconds) {
		return NewPerformanceSecond(second, 0, 0, 0, 0, 0)
	}
	return p.seconds[second].Report(second)
}

func (p *PerformanceStats) second(second int) *secondTally {
	for len(p.seconds) <= second {
		p.seconds = append(p.seconds, secondTally{})
	}
	return &p.seconds[second]
}
