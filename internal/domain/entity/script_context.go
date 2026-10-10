package entity

type ScriptContext struct {
	Response       *Response
	LocalVariables map[string]any
	Variables      VariableContext
	Request        Request
}
