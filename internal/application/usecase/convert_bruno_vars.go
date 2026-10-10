package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
)

const disabledLinePrefix = "// "

func ConvertBrunoPreRequestVars(variables []entity.BrunoVariable) entity.ScriptTranslation {
	translation := entity.ScriptTranslation{Lines: []string{}, Untranslated: []entity.UntranslatedLine{}}
	for _, variable := range variables {
		line := variableSetLine(variable.Key, jsString(variable.Value))
		translation.Lines = append(translation.Lines, enabledLine(line, variable.Enabled))
	}
	return translation
}

func ConvertBrunoPostResponseVars(variables []entity.BrunoVariable) entity.ScriptTranslation {
	translation := entity.ScriptTranslation{Lines: []string{}, Untranslated: []entity.UntranslatedLine{}}
	for _, variable := range variables {
		expression, translated := translateBrunoExpression(variable.Value)
		line := variableSetLine(variable.Key, expression)
		if variable.Enabled && !translated {
			translation.Lines = append(translation.Lines, untranslatedPrefix+line)
			translation.Untranslated = append(translation.Untranslated, entity.UntranslatedLine{Line: variable.Line, Source: line})
			continue
		}
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
