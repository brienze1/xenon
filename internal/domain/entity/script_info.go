package entity

// ScriptInfo is what pm.info and pm.execution.location describe; an IterationCount of 0 leaves the iteration fields undefined.
type ScriptInfo struct {
	RequestID      string
	RequestName    string
	Current        string
	Location       []string
	Iteration      int
	IterationCount int
}

func NewScriptInfo(requestId, requestName, current string, location []string) ScriptInfo {
	return ScriptInfo{RequestID: requestId, RequestName: requestName, Current: current, Location: append([]string{}, location...), IterationCount: 1}
}

func (s ScriptInfo) ForIteration(iteration, count int) ScriptInfo {
	s.Iteration = iteration
	s.IterationCount = count
	return s
}

func (s ScriptInfo) HasIteration() bool {
	return s.IterationCount > 0
}
