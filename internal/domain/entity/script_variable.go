package entity

type ScriptVariable struct {
	Value    any
	Key      string
	Type     string
	Secret   bool
	Disabled bool
}

func NewScriptVariable(key string, value any) ScriptVariable {
	return ScriptVariable{Key: key, Value: value}
}

func (s ScriptVariable) Text() string {
	return LocalVariableText(s.Value)
}
