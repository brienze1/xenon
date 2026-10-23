package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RunPerformanceSecond struct {
	Second      int     `json:"second"`
	ActiveUsers int     `json:"active_users"`
	Requests    int     `json:"requests"`
	AvgMs       float64 `json:"avg_ms"`
	Errors      int     `json:"errors"`
}

func NewRunPerformanceSecond(second entity.PerformanceSecond) RunPerformanceSecond {
	return RunPerformanceSecond{Second: second.Second, ActiveUsers: second.ActiveUsers, Requests: second.Requests, AvgMs: second.AvgMs, Errors: second.Errors}
}
