package entity

const (
	ApiKeyInHeader = "header"
	ApiKeyInQuery  = "query"
)

type ApiKeyAuth struct {
	Key   string
	Value string
	In    string
}

func DefaultApiKeyAuth() ApiKeyAuth {
	return ApiKeyAuth{In: ApiKeyInHeader}
}

func (a ApiKeyAuth) SendsNothing() bool {
	return a.Key == "" && a.Value == ""
}

func (a ApiKeyAuth) IsInQuery() bool {
	return a.In == ApiKeyInQuery
}

func (a ApiKeyAuth) IsBlank() bool {
	return a.SendsNothing() && !a.IsInQuery()
}

func (a ApiKeyAuth) MapText(transform func(string) string) ApiKeyAuth {
	a.Key = transform(a.Key)
	a.Value = transform(a.Value)
	return a
}

func (a ApiKeyAuth) TextFields(prefix string) []TextField {
	return []TextField{NewTextField(fieldName(prefix, "apikey", "key"), a.Key), NewTextField(fieldName(prefix, "apikey", "value"), a.Value)}
}

func (a ApiKeyAuth) LogValue() any {
	return map[string]any{"key": a.Key, "in": a.In, "hasValue": a.Value != ""}
}
