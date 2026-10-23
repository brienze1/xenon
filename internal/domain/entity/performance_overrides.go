package entity

type PerformanceOverrides struct {
	DataFile        *string
	Profile         LoadProfileOverrides
	VirtualUsers    *int
	DurationSeconds *int
}
