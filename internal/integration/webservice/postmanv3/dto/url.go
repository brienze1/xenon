package dto

type URL struct {
	Raw       string
	Query     KeyValues
	Variables KeyValues
}

func NewURL(raw string, query, variables KeyValues) URL {
	return URL{Raw: raw, Query: query, Variables: variables}
}

func (u URL) toPostmanV21() any {
	if u.Query == nil && u.Variables == nil {
		return u.Raw
	}
	url := map[string]any{"raw": u.Raw}
	setIfPresent(url, "query", u.Query.toPostmanV21())
	setIfPresent(url, "variable", u.Variables.toPathVariables())
	return url
}
