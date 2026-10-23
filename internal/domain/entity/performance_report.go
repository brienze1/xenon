package entity

type PerformanceReport struct {
	PerRequest []RequestReport
	Series     []PerformanceSecond
	Overall    RequestReport
}

func NewPerformanceReport(overall RequestReport, perRequest []RequestReport, series []PerformanceSecond) PerformanceReport {
	return PerformanceReport{Overall: overall, PerRequest: perRequest, Series: series}
}
