package entity

import "strconv"

const (
	runCookieJarSeparator = "#run:"
	runResponseKeyPrefix  = "run:"
)

type RunContext struct {
	Variables       *RunVariables
	IterationData   map[string]any
	RunID           string
	EnvironmentID   string
	CookieJarID     string
	ResponseKey     string
	Acknowledgement SendAcknowledgement
	Iteration       int
	IterationCount  int
	MuteConsole     bool
	CollectConsole  bool
}

func NoRunContext() RunContext {
	return RunContext{IterationCount: 1}
}

func NewRunContext(runId, environmentId, cookieJarId string, variables *RunVariables, acknowledgement SendAcknowledgement, iteration int) RunContext {
	return RunContext{
		RunID:           runId,
		EnvironmentID:   environmentId,
		CookieJarID:     cookieJarId,
		Variables:       variables,
		Acknowledgement: acknowledgement,
		Iteration:       iteration,
	}
}

func RunCookieJarID(localId, runId string) string {
	return localId + runCookieJarSeparator + runId
}

func VirtualUserCookieJarID(runJarId string, user int) string {
	return runJarId + ":" + strconv.Itoa(user)
}

func RunResponseKey(runId string, index int) string {
	return runResponseKeyPrefix + runId + ":" + strconv.Itoa(index)
}

// ForFunctionalRow keeps the response of row index when persisted and collects the console unless logs are off.
func (r RunContext) ForFunctionalRow(settings FunctionalSettings, data IterationData, index int) RunContext {
	r.IterationCount = settings.Iterations
	r.IterationData = data.Row(r.Iteration)
	r.MuteConsole = settings.DisableLogs
	r.CollectConsole = !settings.DisableLogs
	if settings.PersistResponses {
		r.ResponseKey = RunResponseKey(r.RunID, index)
	}
	return r
}

func (r RunContext) ForPerformance(data IterationData) RunContext {
	r.IterationCount = data.IterationCount()
	r.IterationData = data.WrappedRow(r.Iteration)
	r.MuteConsole = true
	return r
}

// ResponseKeyFor is the tab id of a tab or nested send, else the key the run keeps the response under.
func (r RunContext) ResponseKeyFor(tabId string, nested bool) string {
	if tabId != "" || nested {
		return tabId
	}
	return r.ResponseKey
}

func (r RunContext) KeepsResponseFor(tabId string, nested bool) bool {
	return r.ResponseKeyFor(tabId, nested) != ""
}

func (r RunContext) IsRun() bool {
	return r.RunID != ""
}

func (r RunContext) JarID(localId string) string {
	if r.CookieJarID == "" {
		return localId
	}
	return r.CookieJarID
}

func (r RunContext) UsesWorkspaceJar(localId string) bool {
	return r.JarID(localId) == localId
}

func (r RunContext) KeepsVariables() bool {
	return r.Variables == nil
}

func (r RunContext) ScriptInfo(info ScriptInfo) ScriptInfo {
	return info.ForIteration(r.Iteration, r.IterationCount)
}

func (r RunContext) HasIterationData() bool {
	return r.IterationData != nil
}

func (r RunContext) WritesConsole() bool {
	return !r.MuteConsole
}

func (r RunContext) Collects() bool {
	return r.CollectConsole
}
