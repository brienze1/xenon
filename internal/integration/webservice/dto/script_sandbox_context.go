package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const (
	scriptPhasePreRequest = "prerequest"
	scriptPhaseTest       = "test"
)

// ScriptSandboxContext is everything the pm shims are built from for one script.
type ScriptSandboxContext struct {
	Response *ScriptResponse       `json:"response,omitempty"`
	Vault    map[string]string     `json:"vault"`
	Info     ScriptSandboxInfo     `json:"info"`
	Location ScriptSandboxLocation `json:"location"`
	Phase    string                `json:"phase"`
	Scopes   ScriptSandboxScopes   `json:"scopes"`
	Cookies  []ScriptSandboxCookie `json:"cookies"`
	Request  ScriptRequest         `json:"request"`
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
	if context.Response != nil && !script.IsPreRequest() {
		response := NewScriptResponse(*context.Response)
		response.Cookie = NewScriptSandboxCookies(context.ResponseCookies)
		sandbox.Response = &response
	}
	return sandbox
}
