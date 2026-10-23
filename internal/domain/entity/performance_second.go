package entity

type PerformanceSecond struct {
	Second      int
	ActiveUsers int
	Requests    int
	Errors      int
	AvgMs       float64
}
