package entity

type RunPlanRow struct {
	RunRow
	Name       string
	Method     string
	FolderPath []string
	Aliases    []RequestAlias
	Number     int
}

// FlattenRunRows lists the requests under the target in sidebar order, depth first; folder paths are relative to the target.
func FlattenRunRows(target TreeNode) []RunPlanRow {
	rows := flattenRunRows(target.Children, []string{})
	for index := range rows {
		rows[index].Number = index + 1
	}
	return rows
}

func flattenRunRows(nodes []TreeNode, folderPath []string) []RunPlanRow {
	rows := make([]RunPlanRow, 0)
	for _, node := range nodes {
		switch {
		case node.Ignored || !node.IsItem():
			continue
		case node.IsRequest():
			rows = append(rows, RunPlanRow{
				RunRow:     RunRow{RequestID: node.ID, Selected: true},
				Name:       node.Name,
				Method:     node.Method,
				FolderPath: append([]string{}, folderPath...),
				Aliases:    []RequestAlias{},
			})
		case node.IsContainer():
			rows = append(rows, flattenRunRows(node.Children, append(append([]string{}, folderPath...), node.Name))...)
		}
	}
	return rows
}

func RunPlanRowOf(rows []RunPlanRow, requestId string) (RunPlanRow, bool) {
	for _, row := range rows {
		if row.RequestID == requestId {
			return row, true
		}
	}
	return RunPlanRow{}, false
}

func (r RunPlanRow) For(row RunRow) RunPlanRow {
	r.RunRow = row
	if r.FolderPath == nil {
		r.FolderPath = []string{}
	}
	return r
}
