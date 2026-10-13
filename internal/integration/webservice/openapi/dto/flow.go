package dto

type Flow struct {
	Name             string `yaml:"-"`
	AuthorizationUrl string `yaml:"authorizationUrl,omitempty"`
	TokenUrl         string `yaml:"tokenUrl,omitempty"`
	Scopes           Scopes `yaml:"scopes"`
}

func (f Flow) MarshalYAML() (any, error) {
	return mappingOfFields(text("authorizationUrl", f.AuthorizationUrl), text("tokenUrl", f.TokenUrl), always("scopes", f.Scopes))
}
