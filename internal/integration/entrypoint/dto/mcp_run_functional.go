package dto

import "github.com/brienze1/xenon/internal/domain/entity"

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

func (m *MCPRunFunctional) ToEntity() entity.FunctionalOverrides {
	if m == nil {
		return entity.FunctionalOverrides{}
	}
	return entity.FunctionalOverrides{
		DataFile:           m.DataFile,
		Iterations:         m.Iterations,
		DelayMs:            m.DelayMs,
		PersistResponses:   m.PersistResponses,
		DisableLogs:        m.DisableLogs,
		StopOnError:        m.StopOnError,
		KeepVariableValues: m.KeepVariableValues,
		SkipStoredCookies:  m.SkipStoredCookies,
		SaveCookies:        m.SaveCookies,
	}
}
