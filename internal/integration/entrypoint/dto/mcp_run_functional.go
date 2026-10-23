package dto

type MCPRunFunctional struct {
	Iterations         *int    `json:"iterations" description:"Iterations; one per data row when a data_file comes without it."`
	DataFile           *string `json:"data_file" description:"CSV or JSON data file path."`
	DelayMs            *int    `json:"delay_ms" description:"Delay between requests."`
	PersistResponses   *bool   `json:"persist_responses" description:"Keep the responses."`
	DisableLogs        *bool   `json:"disable_logs" description:"No console output."`
	StopOnError        *bool   `json:"stop_on_error" description:"Stop at the first error."`
	KeepVariableValues *bool   `json:"keep_variable_values" description:"Save the variables scripts set."`
	SkipStoredCookies  *bool   `json:"skip_stored_cookies" description:"Start without stored cookies."`
	SaveCookies        *bool   `json:"save_cookies" description:"Save the cookies at the end."`
}

func (m *MCPRunFunctional) Over(settings RunFunctionalSettings) RunFunctionalSettings {
	if m == nil {
		return settings
	}
	if m.DataFile != nil {
		settings.DataFile = *m.DataFile
	}
	if m.DataFile != nil && *m.DataFile != "" {
		settings.Iterations = 0
	}
	settings.Iterations = valueOr(m.Iterations, settings.Iterations)
	settings.DelayMs = valueOr(m.DelayMs, settings.DelayMs)
	settings.PersistResponses = valueOr(m.PersistResponses, settings.PersistResponses)
	settings.DisableLogs = valueOr(m.DisableLogs, settings.DisableLogs)
	settings.StopOnError = valueOr(m.StopOnError, settings.StopOnError)
	settings.KeepVariableValues = valueOr(m.KeepVariableValues, settings.KeepVariableValues)
	settings.SkipStoredCookies = valueOr(m.SkipStoredCookies, settings.SkipStoredCookies)
	settings.SaveCookies = valueOr(m.SaveCookies, settings.SaveCookies)
	return settings
}

func valueOr[T any](value *T, fallback T) T {
	if value == nil {
		return fallback
	}
	return *value
}
