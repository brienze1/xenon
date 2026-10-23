package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RunPerformanceMetrics struct {
	ActiveUsers int     `json:"active_users"`
	Requests    int     `json:"requests"`
	AvgMs       float64 `json:"avg_ms"`
	Throughput  float64 `json:"throughput"`
	ErrorRate   float64 `json:"error_rate"`
}

func NewRunPerformanceMetrics(metrics *entity.PerformanceMetrics) *RunPerformanceMetrics {
	if metrics == nil {
		return nil
	}
	return &RunPerformanceMetrics{ActiveUsers: metrics.ActiveUsers, Requests: metrics.Requests, AvgMs: metrics.AvgMs, Throughput: metrics.Throughput, ErrorRate: metrics.ErrorRate}
}
