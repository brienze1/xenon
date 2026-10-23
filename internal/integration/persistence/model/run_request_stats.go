package model

import (
	"maps"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type RunRequestStats struct {
	StatusCounts map[int]int `json:"status_counts"`
	RequestId    string      `json:"request_id"`
	Name         string      `json:"name"`
	Count        int         `json:"count"`
	AvgMs        float64     `json:"avg_ms"`
	P90Ms        float64     `json:"p90_ms"`
	P95Ms        float64     `json:"p95_ms"`
	P99Ms        float64     `json:"p99_ms"`
	MinMs        float64     `json:"min_ms"`
	MaxMs        float64     `json:"max_ms"`
	Throughput   float64     `json:"throughput"`
	ErrorRate    float64     `json:"error_rate"`
}

func NewRunRequestStats(report entity.RequestReport) RunRequestStats {
	return RunRequestStats{
		StatusCounts: maps.Clone(report.StatusCounts),
		RequestId:    report.RequestID,
		Name:         report.Name,
		Count:        report.Count,
		AvgMs:        report.AvgMs,
		P90Ms:        report.P90Ms,
		P95Ms:        report.P95Ms,
		P99Ms:        report.P99Ms,
		MinMs:        report.MinMs,
		MaxMs:        report.MaxMs,
		Throughput:   report.Throughput,
		ErrorRate:    report.ErrorRate,
	}
}

func (r *RunRequestStats) ToEntity() entity.RequestReport {
	statusCounts := maps.Clone(r.StatusCounts)
	if statusCounts == nil {
		statusCounts = map[int]int{}
	}
	return entity.RequestReport{
		StatusCounts: statusCounts,
		RequestID:    r.RequestId,
		Name:         r.Name,
		Count:        r.Count,
		AvgMs:        r.AvgMs,
		P90Ms:        r.P90Ms,
		P95Ms:        r.P95Ms,
		P99Ms:        r.P99Ms,
		MinMs:        r.MinMs,
		MaxMs:        r.MaxMs,
		Throughput:   r.Throughput,
		ErrorRate:    r.ErrorRate,
	}
}
