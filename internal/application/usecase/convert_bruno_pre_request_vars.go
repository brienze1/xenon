package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

const disabledLinePrefix = "// "

func ConvertBrunoPreRequestVars(variables []entity.BrunoVariable) entity.ScriptTranslation {
	translation := entity.NewScriptTranslation()
	for _, variable := range variables {
		line := variableSetLine(variable.Key, jsString(variable.Value))
		translation.Lines = append(translation.Lines, enabledLine(line, variable.Enabled))
	}
	return translation
}

func variableSetLine(key, value string) string {
	return "pm.variables.set(" + jsString(key) + ", " + value + ");"
}

func enabledLine(line string, enabled bool) string {
	if enabled {
		return line
	}
	return disabledLinePrefix + line
}
