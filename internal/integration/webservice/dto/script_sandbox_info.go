package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ScriptSandboxInfo struct {
	Iteration      *int   `json:"iteration,omitempty"`
	IterationCount *int   `json:"iterationCount,omitempty"`
	RequestName    string `json:"requestName,omitempty"`
	RequestId      string `json:"requestId,omitempty"`
}

func NewScriptSandboxInfo(info entity.ScriptInfo) ScriptSandboxInfo {
	sandbox := ScriptSandboxInfo{RequestName: info.RequestName, RequestId: info.RequestID}
	if info.HasIteration() {
		iteration := info.Iteration
		count := info.IterationCount
		sandbox.Iteration = &iteration
		sandbox.IterationCount = &count
	}
	return sandbox
}
