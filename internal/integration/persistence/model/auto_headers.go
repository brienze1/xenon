package model

type AutoHeaders struct {
	Disabled []string `yaml:"disabled" json:"disabled"`
}

func NewAutoHeaders(disabled []string) AutoHeaders {
	return AutoHeaders{Disabled: nonNil(disabled)}
}

func (a *AutoHeaders) ToEntity() []string {
	return nonNil(a.Disabled)
}
