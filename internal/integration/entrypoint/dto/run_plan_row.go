package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RunPlanRow struct {
	RequestId  string         `json:"request_id"`
	AliasId    string         `json:"alias_id"`
	Name       string         `json:"name"`
	Method     string         `json:"method"`
	FolderPath []string       `json:"folder_path"`
	Aliases    []RunPlanAlias `json:"aliases"`
	Number     int            `json:"number"`
	Selected   bool           `json:"selected"`
}

func NewRunPlanRow(row entity.RunPlanRow) RunPlanRow {
	aliases := make([]RunPlanAlias, 0, len(row.Aliases))
	for _, alias := range row.Aliases {
		aliases = append(aliases, NewRunPlanAlias(alias))
	}
	return RunPlanRow{
		Number:     row.Number,
		RequestId:  row.RequestID,
		AliasId:    row.AliasID,
		Selected:   row.Selected,
		Name:       row.Name,
		Method:     row.Method,
		FolderPath: append([]string{}, row.FolderPath...),
		Aliases:    aliases,
	}
}
