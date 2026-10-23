package entity

import "strconv"

const (
	runCookieJarSeparator = "#run:"
	runResponseKeyPrefix  = "run:"
)

// RunContext carries what a collection run changes in the send pipeline; NoRunContext is a single send.
type RunContext struct {
	Collect         func(ConsoleEntry)
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
	Performance     bool
}

func NoRunContext() RunContext {
	return RunContext{IterationCount: 1}
}

func RunCookieJarID(localId, runId string) string {
	return localId + runCookieJarSeparator + runId
}

func RunResponseKey(runId string, index int) string {
	return runResponseKeyPrefix + runId + ":" + strconv.Itoa(index)
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

func (r RunContext) Collects() bool {
	return r.Collect != nil
}
