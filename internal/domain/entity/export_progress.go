package entity

const (
	exportProgressThreshold = 200
	exportProgressStep      = 50
)

type ExportProgress struct {
	Done  int
	Total int
}

func (e ExportProgress) IsMilestone() bool {
	return e.Total > exportProgressThreshold && e.Done > 0 && e.Done < e.Total && e.Done%exportProgressStep == 0
}
