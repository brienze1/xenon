package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

func ConvertBrunoPostResponseVars(variables []entity.BrunoVariable) entity.ScriptTranslation {
	translation := entity.NewScriptTranslation()
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
