package dto

const (
	noAuthType  = "noauth"
	inheritType = "inherit"
)

type AuthEntry struct {
	Extra       map[string]any `yaml:",inline"`
	ID          any            `yaml:"id"`
	Name        string         `yaml:"name"`
	Type        string         `yaml:"type"`
	Credentials KeyValues      `yaml:"credentials"`
}

func (a *AuthEntry) toPostmanV21() map[string]any {
	auth := copied(a.Extra)
	auth["type"] = a.Type
	if a.Type == noAuthType {
		return auth
	}
	rows := make([]any, 0, len(a.Credentials))
	for index := range a.Credentials {
		rows = append(rows, a.Credentials[index].toCredential())
	}
	auth[a.Type] = rows
	return auth
}
