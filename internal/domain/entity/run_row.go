package entity

type RunRow struct {
	RequestID string
	AliasID   string
	Selected  bool
}

func NewSelectedRunRow(requestId string) RunRow {
	return RunRow{RequestID: requestId, Selected: true}
}

func (r RunRow) IsSelected() bool {
	return r.Selected
}
