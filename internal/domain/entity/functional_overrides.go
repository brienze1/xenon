package entity

type FunctionalOverrides struct {
	DataFile           *string
	Iterations         *int
	DelayMs            *int
	PersistResponses   *bool
	DisableLogs        *bool
	StopOnError        *bool
	KeepVariableValues *bool
	SkipStoredCookies  *bool
	SaveCookies        *bool
}
