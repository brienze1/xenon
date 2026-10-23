package model

import "github.com/brienze1/xenon/internal/domain/entity"

type FunctionalSettings struct {
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

func NewFunctionalSettings(settings entity.FunctionalSettings) FunctionalSettings {
	return FunctionalSettings{
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

func (f *FunctionalSettings) ToEntity() entity.FunctionalSettings {
	return entity.FunctionalSettings{
		DataFile:           f.DataFile,
		Iterations:         f.Iterations,
		DelayMs:            f.DelayMs,
		PersistResponses:   f.PersistResponses,
		DisableLogs:        f.DisableLogs,
		StopOnError:        f.StopOnError,
		KeepVariableValues: f.KeepVariableValues,
		SkipStoredCookies:  f.SkipStoredCookies,
		SaveCookies:        f.SaveCookies,
	}
}
