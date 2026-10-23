package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const (
	scriptPhasePreRequest = "prerequest"
	scriptPhaseTest       = "test"
)

type ScriptSandboxContext struct {
	Response      *ScriptResponse       `json:"response,omitempty"`
	Vault         map[string]string     `json:"vault"`
	IterationData map[string]any        `json:"iterationData"`
	Info          ScriptSandboxInfo     `json:"info"`
	Location      ScriptSandboxLocation `json:"location"`
	Phase         string                `json:"phase"`
	Scopes        ScriptSandboxScopes   `json:"scopes"`
	Cookies       []ScriptSandboxCookie `json:"cookies"`
	Request       ScriptRequest         `json:"request"`
}

func NewScriptSandboxContext(script entity.ScriptRef, context entity.ScriptContext) ScriptSandboxContext {
	sandbox := ScriptSandboxContext{
		Phase:    scriptPhaseTest,
		Request:  NewScriptRequest(context.Request),
		Scopes:   NewScriptSandboxScopes(context.Scopes),
		Vault:    context.Vault,
		Info:     NewScriptSandboxInfo(context.Info),
		Location: NewScriptSandboxLocation(context.Info),
		Cookies:  NewScriptSandboxCookies(context.Cookies),
	}
	if script.IsPreRequest() {
		sandbox.Phase = scriptPhasePreRequest
	}
	if sandbox.Vault == nil {
		sandbox.Vault = map[string]string{}
	}
	if context.IterationData != nil {
		sandbox.IterationData = context.IterationData
	} else {
		sandbox.IterationData = map[string]any{}
	}
	if context.Response != nil && !script.IsPreRequest() {
		response := NewScriptResponse(*context.Response)
		response.Cookie = NewScriptSandboxCookies(context.ResponseCookies)
		sandbox.Response = &response
	}
	return sandbox
}
