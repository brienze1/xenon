package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/clashresolution"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

func PlanEnvironmentImport(imported entity.ImportVariables, existing []entity.Environment, resolution enums.ClashResolution) entity.ImportVariablesWrite {
	environment := imported.Environment
	match, clashes := sameNamedEnvironment(environment.Name, existing)
	write := entity.ImportVariablesWrite{Secrets: importedSecrets(imported)}
	switch {
	case !clashes:
		environment = withEnvironmentFileName(environment, existing)
	case resolution == clashresolution.Replace:
		environment = environment.WithFileName(match.FileName)
		write.Replaced = &match
	case resolution == clashresolution.KeepBoth:
		environment.Name = ResolveClashName(environment.Name, environmentNames(existing))
		environment = withEnvironmentFileName(environment, existing)
	default:
		return entity.ImportVariablesWrite{}
	}
	write.Environment = &environment
	return write
}

func PlanGlobalsImport(imported entity.ImportVariables, workspace entity.Workspace, current map[string]string, resolution enums.ClashResolution) entity.ImportVariablesWrite {
	merged, kept := MergeImportVariables(workspace.Variables, imported.Rows(), resolution)
	secrets := make(map[string]string, len(current))
	for key, value := range current {
		secrets[key] = value
	}
	for _, row := range imported.Rows() {
		if _, clashes := variableIndex(workspace.Variables, row.Key); clashes && resolution != clashresolution.Replace {
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

func importedSecrets(imported entity.ImportVariables) map[string]string {
	secrets := map[string]string{}
	for _, row := range imported.Rows() {
		if value, found := imported.SecretOf(row); found {
			secrets[row.Key] = value
		}
	}
	return secrets
}

func withEnvironmentFileName(environment entity.Environment, existing []entity.Environment) entity.Environment {
	siblings := entity.NewEnvironmentList(existing, "").SiblingFileNames(environment)
	return environment.WithFileName(DeriveFileName(environment.Name, itemkind.Environment, siblings))
}

func environmentNames(existing []entity.Environment) []string {
	names := make([]string, 0, len(existing))
	for _, environment := range existing {
		names = append(names, environment.Name)
	}
	return names
}
