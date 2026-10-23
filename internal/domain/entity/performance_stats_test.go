package entity

import (
	"sync"
	"testing"
	"time"
)

func TestPerformanceStatsReportsOverallPerRequestAndSeries(t *testing.T) {
	stats := NewPerformanceStats([]RunPlanRow{{RunRow: RunRow{RequestID: "a"}, Name: "A"}, {RunRow: RunRow{RequestID: "b"}, Name: "B"}})
	stats.Observe(0, 2)
	stats.Observe(0, 1)
	stats.Observe(1, 3)
	var waiting sync.WaitGroup
	for index := range 10 {
		waiting.Add(1)
		go func() {
			defer waiting.Done()
			stats.Record(PerformanceSample{Row: index % 2, Second: index / 5, Status: 200, Latency: 10 * time.Millisecond})
		}()
	}
	waiting.Wait()
	stats.Record(PerformanceSample{Row: 1, Second: 1, Status: 500, Latency: 30 * time.Millisecond})
	stats.Record(PerformanceSample{Row: 1, Second: 1})

	report := stats.Report(4*time.Second, 3)
	overall := report.Overall
	if overall.Count != 12 || overall.StatusCounts[200] != 10 || overall.StatusCounts[500] != 1 || overall.Throughput != 3 {
		t.Errorf("overall = %+v", overall)
	}
	if overall.ErrorRate != 2.0/12 || overall.MinMs != 10 || overall.MaxMs != 30 {
		t.Errorf("overall = %+v", overall)
	}
	if len(report.PerRequest) != 2 || report.PerRequest[0].Count != 5 || report.PerRequest[1].Count != 7 || report.PerRequest[1].Name != "B" {
		t.Errorf("per request = %+v", report.PerRequest)
	}
	if len(report.Series) != 3 {
		t.Fatalf("series = %+v", report.Series)
	}
	if first := report.Series[0]; first.ActiveUsers != 2 || first.Requests != 5 || first.AvgMs != 10 || first.Errors != 0 {
		t.Errorf("second 0 = %+v", first)
	}
	if second := report.Series[1]; second.ActiveUsers != 3 || second.Requests != 7 || second.Errors != 2 {
		t.Errorf("second 1 = %+v", second)
	}
	if third := report.Series[2]; third.Second != 2 || third.Requests != 0 {
		t.Errorf("second 2 = %+v", third)
	}
}

func TestPerformanceStatsMetrics(t *testing.T) {
	stats := NewPerformanceStats([]RunPlanRow{{RunRow: RunRow{RequestID: "a"}, Name: "A"}})
	stats.Record(PerformanceSample{Status: 200, Latency: 20 * time.Millisecond})
	stats.Record(PerformanceSample{Status: 404, Latency: 40 * time.Millisecond})
	metrics := stats.Metrics(4, 2*time.Second)
	if metrics != (PerformanceMetrics{ActiveUsers: 4, Requests: 2, AvgMs: 30, Throughput: 1, ErrorRate: 0.5}) {
		t.Errorf("metrics = %+v", metrics)
	}
}

func TestPerformanceSampleErrors(t *testing.T) {
	for _, testCase := range []struct {
		status int
		error  bool
	}{{0, true}, {200, false}, {399, false}, {400, true}, {503, true}} {
		if got := (PerformanceSample{Status: testCase.status}).IsError(); got != testCase.error {
			t.Errorf("status %d error = %v", testCase.status, got)
		}
	}
}
