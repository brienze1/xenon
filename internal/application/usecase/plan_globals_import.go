package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
)

func PlanGlobalsImport(imported entity.ImportVariables, workspace entity.Workspace, current map[string]string, resolution enums.ClashResolution) entity.ImportVariablesWrite {
	merged, kept := entity.MergeImportedVariables(workspace.Variables, imported.Rows(), resolution)
	secrets := MergeImportedSecrets(workspace.Variables, imported.Rows(), current, imported.Secrets, resolution)
	globals := workspace.WithVariables(merged).WithPostmanGlobals(imported.Environment.Postman)
	return entity.ImportVariablesWrite{Globals: &globals, Secrets: secrets, KeptVariables: kept, Added: len(merged) - len(workspace.Variables)}
}
