package dto

const (
	secretType = "secret"
	cliMask    = "********"
)

type Variable struct {
	Extra       map[string]any `yaml:",inline"`
	Value       any            `yaml:"value"`
	Enabled     *bool          `yaml:"enabled"`
	Source      any            `yaml:"source"`
	Key         string         `yaml:"key"`
	Description string         `yaml:"description"`
	Type        string         `yaml:"type"`
	Disabled    bool           `yaml:"disabled"`
	Secret      bool           `yaml:"secret"`
}

func (v *Variable) isSecret() bool {
	return v.Secret || v.Type == secretType || v.Source != nil
}

func (v *Variable) value() string {
	if v.isSecret() && (v.Source != nil || text(v.Value) == cliMask) {
		return ""
	}
	return text(v.Value)
}

func (v *Variable) toCollectionVariable() map[string]any {
	row := copied(v.Extra)
	row["key"], row["value"] = v.Key, v.value()
	if v.Disabled {
		row["disabled"] = true
	}
	if v.Description != "" {
		row["description"] = v.Description
	}
	if v.isSecret() {
		row["type"] = secretType
	} else if v.Type != "" {
		row["type"] = v.Type
	}
	return row
}

func (v *Variable) toEnvironmentValue(conversion *Conversion, path string) map[string]any {
	row := copied(v.Extra)
	row["key"], row["value"] = v.Key, v.value()
	if v.Enabled != nil {
		row["enabled"] = *v.Enabled
	}
	if v.Description != "" {
		row["description"] = v.Description
	}
	if v.isSecret() {
		row["type"] = secretType
	} else if v.Type != "" {
		row["type"] = v.Type
	}
	if v.Source != nil {
		row["source"] = v.Source
		conversion.linkedSecret(v.Key, path)
	}
	return row
}
