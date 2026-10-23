package dto

import (
	"strconv"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type RunRequestStats struct {
	StatusCounts map[string]int `json:"status_counts"`
	RequestId    string         `json:"request_id"`
	Name         string         `json:"name"`
	Count        int            `json:"count"`
	AvgMs        float64        `json:"avg_ms"`
	P90Ms        float64        `json:"p90_ms"`
	P95Ms        float64        `json:"p95_ms"`
	P99Ms        float64        `json:"p99_ms"`
	MinMs        float64        `json:"min_ms"`
	MaxMs        float64        `json:"max_ms"`
	Throughput   float64        `json:"throughput"`
	ErrorRate    float64        `json:"error_rate"`
}

func NewRunRequestStats(report entity.RequestReport) RunRequestStats {
	statusCounts := make(map[string]int, len(report.StatusCounts))
	for status, count := range report.StatusCounts {
		statusCounts[strconv.Itoa(status)] = count
	}
	return RunRequestStats{
		StatusCounts: statusCounts,
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
