package dto

import "github.com/brienze1/xenon/internal/domain/entity"

// ScriptSandboxResult is what the pm shims hand back when a script ends.
type ScriptSandboxResult struct {
	Request *ScriptRequest            `json:"request"`
	Vault   map[string]any            `json:"vault"`
	Scopes  ScriptSandboxScopes       `json:"scopes"`
	Tests   []ScriptSandboxTestResult `json:"tests"`
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
		tests = append(tests, test.ToEntity())
	}
	return tests
}
