package model

import "github.com/brienze1/xenon/internal/domain/entity"

type RunPerformanceSecond struct {
	Second      int     `json:"second"`
	ActiveUsers int     `json:"active_users"`
	Requests    int     `json:"requests"`
	Errors      int     `json:"errors"`
	AvgMs       float64 `json:"avg_ms"`
}

func NewRunPerformanceSecond(second entity.PerformanceSecond) RunPerformanceSecond {
	return RunPerformanceSecond(second)
}

func (r *RunPerformanceSecond) ToEntity() entity.PerformanceSecond {
	return entity.PerformanceSecond(*r)
}
