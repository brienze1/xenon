package entity

// ScriptScope is one variable scope as scripts see it: ordered values, typed as the script left them.
type ScriptScope struct {
	Name     string
	Values   []ScriptVariable
	Inactive bool
}

func NewScriptScope(name string, values []ScriptVariable) ScriptScope {
	return ScriptScope{Name: name, Values: append([]ScriptVariable{}, values...)}
}

func NewInactiveScriptScope() ScriptScope {
	return ScriptScope{Values: []ScriptVariable{}, Inactive: true}
}

func (s ScriptScope) Lookup(key string) (ScriptVariable, bool) {
	for index := len(s.Values) - 1; index >= 0; index-- {
		if s.Values[index].Key == key && !s.Values[index].Disabled {
			return s.Values[index], true
		}
	}
	return ScriptVariable{}, false
}

func (s ScriptScope) WithValue(key string, value any) ScriptScope {
	values := make([]ScriptVariable, 0, len(s.Values)+1)
	replaced := false
	for _, variable := range s.Values {
		if variable.Key == key && !variable.Disabled {
			variable.Value = value
			replaced = true
		}
		values = append(values, variable)
	}
	if !replaced {
		values = append(values, NewScriptVariable(key, value))
	}
	s.Values = values
	return s
}

func (s ScriptScope) WithoutValue(key string) ScriptScope {
	values := make([]ScriptVariable, 0, len(s.Values))
	for _, variable := range s.Values {
		if variable.Key != key {
			values = append(values, variable)
		}
	}
	s.Values = values
	return s
}

func (s ScriptScope) ValueMap() map[string]any {
	values := make(map[string]any, len(s.Values))
	for _, variable := range s.Values {
		if !variable.Disabled {
			values[variable.Key] = variable.Value
		}
	}
	return values
}

func (s ScriptScope) TextMap() map[string]string {
	values := make(map[string]string, len(s.Values))
	for _, variable := range s.Values {
		if !variable.Disabled {
			values[variable.Key] = variable.Text()
		}
	}
	return values
}

// ChangesFrom lists what a script did to the scope, in the scope's order: values whose text changed or appeared, and names it removed.
func (s ScriptScope) ChangesFrom(previous ScriptScope) ([]KeyValue, []string) {
	before := previous.TextMap()
	after := s.TextMap()
	set := make([]KeyValue, 0)
	seen := make(map[string]bool)
	for _, variable := range s.Values {
		text, found := after[variable.Key]
		if !found || variable.Disabled || seen[variable.Key] {
			continue
		}
		seen[variable.Key] = true
		if old, existed := before[variable.Key]; !existed || old != text {
			set = append(set, NewKeyValue(variable.Key, text))
		}
	}
	unset := make([]string, 0)
	for _, variable := range previous.Values {
		if _, kept := after[variable.Key]; !kept && !variable.Disabled && !containsKey(unset, variable.Key) {
			unset = append(unset, variable.Key)
		}
	}
	return set, unset
}

func containsKey(keys []string, key string) bool {
	for _, existing := range keys {
		if existing == key {
			return true
		}
	}
	return false
}
