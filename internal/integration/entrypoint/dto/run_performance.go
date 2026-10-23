package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RunPerformance struct {
	PerRequest []RunRequestStats      `json:"per_request"`
	Series     []RunPerformanceSecond `json:"series"`
	Overall    RunRequestStats        `json:"overall"`
}

func NewRunPerformance(report *entity.PerformanceReport) *RunPerformance {
	if report == nil {
		return nil
	}
	perRequest := make([]RunRequestStats, 0, len(report.PerRequest))
	for _, request := range report.PerRequest {
		perRequest = append(perRequest, NewRunRequestStats(request))
	}
	series := make([]RunPerformanceSecond, 0, len(report.Series))
	for _, second := range report.Series {
		series = append(series, NewRunPerformanceSecond(second))
	}
	return &RunPerformance{Overall: NewRunRequestStats(report.Overall), PerRequest: perRequest, Series: series}
}
