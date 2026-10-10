package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/teststatus"
)

const (
	scriptPhasePreRequest = "prerequest"
	scriptPhaseTest       = "test"
)

type ScriptSandboxVariable struct {
	Value    any    `json:"value"`
	Key      string `json:"key"`
	Type     string `json:"type,omitempty"`
	Secret   bool   `json:"secret,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
}

type ScriptSandboxScope struct {
	Active *bool                   `json:"active,omitempty"`
	Name   string                  `json:"name,omitempty"`
	Values []ScriptSandboxVariable `json:"values"`
}

// ScriptSandboxScopes uses the property names postman-sandbox gives the scopes.
type ScriptSandboxScopes struct {
	Environment         ScriptSandboxScope `json:"environment"`
	CollectionVariables ScriptSandboxScope `json:"collectionVariables"`
	Globals             ScriptSandboxScope `json:"globals"`
	Local               ScriptSandboxScope `json:"local"`
}

type ScriptSandboxInfo struct {
	Iteration      *int   `json:"iteration,omitempty"`
	IterationCount *int   `json:"iterationCount,omitempty"`
	RequestName    string `json:"requestName,omitempty"`
	RequestId      string `json:"requestId,omitempty"`
}

type ScriptSandboxLocation struct {
	Current string   `json:"current,omitempty"`
	Path    []string `json:"path"`
}

type ScriptSandboxCookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain,omitempty"`
	Path     string `json:"path,omitempty"`
	Expires  string `json:"expires,omitempty"`
	MaxAge   int    `json:"maxAge,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
	HttpOnly bool   `json:"httpOnly,omitempty"`
	HostOnly bool   `json:"hostOnly,omitempty"`
}

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

type ScriptSandboxTest struct {
	Error   *string `json:"error"`
	Name    string  `json:"name"`
	Index   int     `json:"index"`
	Passed  bool    `json:"passed"`
	Skipped bool    `json:"skipped"`
	Async   bool    `json:"async"`
}

// ScriptSandboxResult is what the pm shims hand back when a script ends.
type ScriptSandboxResult struct {
	Request *ScriptRequest      `json:"request"`
	Vault   map[string]any      `json:"vault"`
	Scopes  ScriptSandboxScopes `json:"scopes"`
	Tests   []ScriptSandboxTest `json:"tests"`
}

func NewScriptSandboxContext(script entity.ScriptRef, context entity.ScriptContext) ScriptSandboxContext {
	sandbox := ScriptSandboxContext{
		Phase:    scriptPhaseTest,
		Request:  NewScriptRequest(context.Request),
		Scopes:   NewScriptSandboxScopes(context.Scopes),
		Vault:    context.Vault,
		Info:     newScriptSandboxInfo(context.Info),
		Location: ScriptSandboxLocation{Path: append([]string{}, context.Info.Location...), Current: context.Info.Current},
	}
	if script.IsPreRequest() {
		sandbox.Phase = scriptPhasePreRequest
	}
	if sandbox.Vault == nil {
		sandbox.Vault = map[string]string{}
	}
	if context.Response != nil && !script.IsPreRequest() {
		response := NewScriptResponse(*context.Response)
		response.Cookie = newScriptSandboxCookies(context.ResponseCookies)
		sandbox.Response = &response
	}
	sandbox.Cookies = newScriptSandboxCookies(context.Cookies)
	return sandbox
}

func newScriptSandboxCookies(cookies []entity.ScriptCookie) []ScriptSandboxCookie {
	converted := make([]ScriptSandboxCookie, 0, len(cookies))
	for _, cookie := range cookies {
		converted = append(converted, ScriptSandboxCookie(cookie))
	}
	return converted
}

func newScriptSandboxInfo(info entity.ScriptInfo) ScriptSandboxInfo {
	sandbox := ScriptSandboxInfo{RequestName: info.RequestName, RequestId: info.RequestID}
	if info.HasIteration() {
		iteration := 0
		count := info.IterationCount
		sandbox.Iteration = &iteration
		sandbox.IterationCount = &count
	}
	return sandbox
}

func NewScriptSandboxScopes(scopes entity.ScriptScopes) ScriptSandboxScopes {
	environment := newScriptSandboxScope(scopes.Environment)
	if scopes.Environment.Inactive {
		active := false
		environment.Active = &active
	}
	return ScriptSandboxScopes{
		Environment:         environment,
		CollectionVariables: newScriptSandboxScope(scopes.Collection),
		Globals:             newScriptSandboxScope(scopes.Globals),
		Local:               newScriptSandboxScope(scopes.Local),
	}
}

func newScriptSandboxScope(scope entity.ScriptScope) ScriptSandboxScope {
	values := make([]ScriptSandboxVariable, 0, len(scope.Values))
	for _, variable := range scope.Values {
		values = append(values, ScriptSandboxVariable{Key: variable.Key, Value: variable.Value, Type: variable.Type, Secret: variable.Secret, Disabled: variable.Disabled})
	}
	return ScriptSandboxScope{Name: scope.Name, Values: values}
}

// ToEntity keeps what the script cannot change about a scope (its name, whether it is active) from the scopes it started with.
func (s ScriptSandboxScopes) ToEntity(started entity.ScriptScopes) entity.ScriptScopes {
	return entity.ScriptScopes{
		Environment: s.Environment.toEntity(started.Environment),
		Collection:  s.CollectionVariables.toEntity(started.Collection),
		Globals:     s.Globals.toEntity(started.Globals),
		Local:       s.Local.toEntity(started.Local),
	}
}

func (s ScriptSandboxScope) toEntity(started entity.ScriptScope) entity.ScriptScope {
	values := make([]entity.ScriptVariable, 0, len(s.Values))
	for _, variable := range s.Values {
		values = append(values, entity.ScriptVariable{Key: variable.Key, Value: variable.Value, Type: variable.Type, Secret: variable.Secret, Disabled: variable.Disabled})
	}
	return entity.ScriptScope{Name: started.Name, Values: values, Inactive: started.Inactive}
}

func (s ScriptSandboxResult) VaultEntity() map[string]string {
	vault := make(map[string]string, len(s.Vault))
	for key, value := range s.Vault {
		if value != nil {
			vault[key] = entity.LocalVariableText(value)
		}
	}
	return vault
}

func (s ScriptSandboxResult) TestsEntity() []entity.TestResult {
	tests := make([]entity.TestResult, 0, len(s.Tests))
	for _, test := range s.Tests {
		tests = append(tests, test.toEntity())
	}
	return tests
}

func (s ScriptSandboxTest) toEntity() entity.TestResult {
	result := entity.TestResult{Name: s.Name, Index: s.Index, Async: s.Async, Status: teststatus.Passed}
	switch {
	case s.Skipped:
		result.Status = teststatus.Skipped
	case !s.Passed:
		result.Status = teststatus.Failed
		if s.Error != nil {
			result.ErrorMessage = *s.Error
		}
	}
	return result
}
