package dto

type EnvironmentVariable struct {
	Key     string
	Value   string
	Enabled bool
	Secret  bool
}

func NewEnvironmentVariable(key, value string, enabled, secret bool) EnvironmentVariable {
	return EnvironmentVariable{Key: key, Value: value, Enabled: enabled, Secret: secret}
}
