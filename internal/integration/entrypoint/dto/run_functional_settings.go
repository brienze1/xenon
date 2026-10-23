package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RunFunctionalSettings struct {
	DataFile           string `json:"data_file"`
	Iterations         int    `json:"iterations"`
	DelayMs            int    `json:"delay_ms"`
	PersistResponses   bool   `json:"persist_responses"`
	DisableLogs        bool   `json:"disable_logs"`
	StopOnError        bool   `json:"stop_on_error"`
	KeepVariableValues bool   `json:"keep_variable_values"`
	SkipStoredCookies  bool   `json:"skip_stored_cookies"`
	SaveCookies        bool   `json:"save_cookies"`
}

func NewRunFunctionalSettings(settings entity.FunctionalSettings) RunFunctionalSettings {
	return RunFunctionalSettings{
		DataFile:           settings.DataFile,
		Iterations:         settings.Iterations,
		DelayMs:            settings.DelayMs,
		PersistResponses:   settings.PersistResponses,
		DisableLogs:        settings.DisableLogs,
		StopOnError:        settings.StopOnError,
		KeepVariableValues: settings.KeepVariableValues,
		SkipStoredCookies:  settings.SkipStoredCookies,
		SaveCookies:        settings.SaveCookies,
	}
}

func (r *RunFunctionalSettings) ToEntity() entity.FunctionalSettings {
	return entity.FunctionalSettings{
		DataFile:           r.DataFile,
		Iterations:         r.Iterations,
		DelayMs:            r.DelayMs,
		PersistResponses:   r.PersistResponses,
		DisableLogs:        r.DisableLogs,
		StopOnError:        r.StopOnError,
		KeepVariableValues: r.KeepVariableValues,
		SkipStoredCookies:  r.SkipStoredCookies,
		SaveCookies:        r.SaveCookies,
	}
}
