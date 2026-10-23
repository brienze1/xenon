package entity

type ScriptEffects struct {
	NextRequest *NextRequest
	Vault       map[string]string
	Scopes      ScriptScopes
	Logs        []ScriptLog
	Notices     []ScriptLog
	Nested      []NestedRequest
	Tests       []TestResult
	Request     Request
}

func (s ScriptEffects) TestsOf(script ScriptRef) []TestResult {
	tests := make([]TestResult, 0, len(s.Tests))
	for _, test := range s.Tests {
		test.Script = script
		tests = append(tests, test)
	}
	return tests
}
