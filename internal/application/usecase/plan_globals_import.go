package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/clashresolution"
)

func PlanGlobalsImport(imported entity.ImportVariables, workspace entity.Workspace, current map[string]string, resolution enums.ClashResolution) entity.ImportVariablesWrite {
	merged, kept := entity.MergeImportedVariables(workspace.Variables, imported.Rows(), resolution)
	secrets := make(map[string]string, len(current))
	for key, value := range current {
		secrets[key] = value
	}
	for _, row := range imported.Rows() {
		if _, clashes := entity.VariableRowIndex(workspace.Variables, row.Key); clashes && resolution != clashresolution.Replace {
			continue
		}
		delete(secrets, row.Key)
		if value, found := imported.SecretOf(row); found {
			secrets[row.Key] = value
		}
	}
	globals := workspace.WithVariables(merged).WithPostmanGlobals(imported.Environment.Postman)
	return entity.ImportVariablesWrite{Globals: &globals, Secrets: secrets, KeptVariables: kept, Added: len(merged) - len(workspace.Variables)}
}
