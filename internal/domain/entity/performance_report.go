package entity

type PerformanceReport struct {
	PerRequest []RequestReport
	Series     []PerformanceSecond
	Overall    RequestReport
}
