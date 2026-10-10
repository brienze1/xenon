package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/clashresolution"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

func PlanEnvironmentImport(imported entity.ImportVariables, existing []entity.Environment, resolution enums.ClashResolution) entity.ImportVariablesWrite {
	environment := imported.Environment
	match, clashes := entity.NamedEnvironment(existing, environment.Name)
	write := entity.ImportVariablesWrite{Secrets: imported.SecretValues()}
	switch {
	case !clashes:
		environment = withEnvironmentFileName(environment, existing)
	case resolution == clashresolution.Replace:
		environment = environment.WithFileName(match.FileName)
		write.Replaced = &match
	case resolution == clashresolution.KeepBoth:
		environment = withEnvironmentFileName(environment.Renamed(entity.ImportedName(environment.Name, environmentNames(existing))), existing)
	default:
		return entity.ImportVariablesWrite{}
	}
	write.Environment = &environment
	return write
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
