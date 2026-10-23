package entity

import (
	"slices"

	"github.com/brienze1/xenon/pkg/errs"
)

type RunPlanRow struct {
	RunRow
	Name       string
	Method     string
	FolderPath []string
	Aliases    []RequestAlias
	Number     int
}

func (r RunPlanRow) MissingRequest() error {
	return errs.NotFoundError("request not found: "+r.RequestID, "RUN-06404")
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

// ReconcileRunRows keeps the remembered order of requests still in the sidebar, adds new ones selected after their nearest present predecessor, and drops aliases that no longer exist.
func ReconcileRunRows(remembered []RunRow, sidebar []RunPlanRow) []RunPlanRow {
	rows := make([]RunPlanRow, 0, len(sidebar))
	for _, row := range remembered {
		if planned, found := RunPlanRowOf(sidebar, row.RequestID); found && !containsRunRow(rows, row.RequestID) {
			rows = append(rows, planned.For(row).withKnownAlias())
		}
	}
	for index, planned := range sidebar {
		if containsRunRow(rows, planned.RequestID) {
			continue
		}
		at := insertionPoint(rows, sidebar[:index])
		rows = slices.Insert(rows, at, planned.For(NewSelectedRunRow(planned.RequestID)))
	}
	for index := range rows {
		rows[index].Number = index + 1
	}
	return rows
}

func (r RunPlanRow) WithAliases(aliases []RequestAlias) RunPlanRow {
	r.Aliases = make([]RequestAlias, 0, len(aliases))
	for _, alias := range aliases {
		r.Aliases = append(r.Aliases, RequestAlias{ID: alias.ID, Name: alias.Name})
	}
	return r
}

func (r RunPlanRow) withKnownAlias() RunPlanRow {
	if IsDefaultAliasID(r.AliasID) {
		return r
	}
	for _, alias := range r.Aliases {
		if alias.ID == r.AliasID {
			return r
		}
	}
	r.AliasID = ""
	return r
}

func insertionPoint(rows []RunPlanRow, preceding []RunPlanRow) int {
	for index := len(preceding) - 1; index >= 0; index-- {
		for at, row := range rows {
			if row.RequestID == preceding[index].RequestID {
				return at + 1
			}
		}
	}
	return 0
}

func containsRunRow(rows []RunPlanRow, requestId string) bool {
	_, found := RunPlanRowOf(rows, requestId)
	return found
}

// FindNextRunRow resolves a setNextRequest target by request id, then by the first row with that name.
func FindNextRunRow(rows []RunPlanRow, target string) (int, bool) {
	for index, row := range rows {
		if row.RequestID == target {
			return index, true
		}
	}
	for index, row := range rows {
		if row.Name == target {
			return index, true
		}
	}
	return 0, false
}

// NextRunPosition is where the iteration continues after the row at current: the next row, the setNextRequest target,
// or past the last row to end it; the target is unknown when it is not a row of the run.
func NextRunPosition(rows []RunPlanRow, current int, next *NextRequest) (int, bool) {
	switch {
	case next == nil:
		return current + 1, true
	case next.Stop:
		return len(rows), true
	}
	if index, found := FindNextRunRow(rows, next.Target); found {
		return index, true
	}
	return len(rows), false
}

func UnknownNextRequestWarning(target string) string {
	return `Couldn't find request "` + target + `" in this run; the iteration ends.`
}
