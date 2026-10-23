package entity

// ScriptEffects is everything one script run leaves behind for the send.
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
