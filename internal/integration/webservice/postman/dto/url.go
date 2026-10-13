package dto

import (
	"encoding/json"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

var urlKeys = []string{"raw", "query", "variable"}

type Url struct {
	Raw      map[string]any `json:"-"`
	Text     string         `json:"raw"`
	Protocol string         `json:"protocol"`
	Port     string         `json:"port"`
	Hash     string         `json:"hash"`
	Host     any            `json:"host"`
	Path     any            `json:"path"`
	Query    []KeyValue     `json:"query"`
	Variable []KeyValue     `json:"variable"`
	IsString bool           `json:"-"`
}

func NewTextUrl(raw string) Url {
	return Url{Text: raw, IsString: true, Query: queryOf(raw)}
}

func (u *Url) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err == nil {
		*u = NewTextUrl(raw)
		return nil
	}
	type plain Url
	if err := json.Unmarshal(data, (*plain)(u)); err != nil {
		return err
	}
	u.Raw = decodeRaw(data)
	return nil
}

func (u *Url) String() string {
	if u.Text != "" || u.IsString {
		return u.Text
	}
	built := joined(u.Host, ".")
	if u.Protocol != "" {
		built = u.Protocol + "://" + built
	}
	if u.Port != "" {
		built += ":" + u.Port
	}
	if path := joined(u.Path, "/"); path != "" {
		built += "/" + path
	}
	if len(u.Query) > 0 {
		pairs := make([]string, 0, len(u.Query))
		for _, row := range u.Query {
			pairs = append(pairs, text(row.Key)+"="+text(row.Value))
		}
		built += "?" + strings.Join(pairs, "&")
	}
	if u.Hash != "" {
		built += "#" + u.Hash
	}
	return built
}

func (u *Url) Params() []entity.KeyValue {
	return keyValuesToEntity(u.Query)
}

func (u *Url) PathVariables() []entity.PathVariable {
	variables := make([]entity.PathVariable, 0, len(u.Variable))
	for index := range u.Variable {
		variables = append(variables, u.Variable[index].ToPathVariable())
	}
	return variables
}

func (u *Url) Passthrough() map[string]any {
	fields := without(u.Raw, urlKeys...)
	fields = withList(fields, "query", keyValuePassthrough(u.Query))
	return withList(fields, "variable", keyValuePassthrough(u.Variable))
}

func joined(value any, separator string) string {
	if parts, isList := value.([]any); isList {
		segments := make([]string, 0, len(parts))
		for _, part := range parts {
			segments = append(segments, text(part))
		}
		return strings.Join(segments, separator)
	}
	return text(value)
}

func queryOf(raw string) []KeyValue {
	_, query, found := strings.Cut(raw, "?")
	if !found {
		return nil
	}
	query, _, _ = strings.Cut(query, "#")
	rows := make([]KeyValue, 0)
	for _, pair := range strings.Split(query, "&") {
		if pair == "" {
			continue
		}
		key, value, _ := strings.Cut(pair, "=")
		rows = append(rows, NewKeyValue(key, value))
	}
	return rows
}

func NewUrl(raw string, params []entity.KeyValue, pathVariables []entity.PathVariable, passthrough map[string]any) Url {
	query := make([]KeyValue, 0, len(params))
	for index, param := range params {
		query = append(query, NewKeyValueRow(param, mapAt(asList(passthrough["query"]), index)))
	}
	variables := make([]KeyValue, 0, len(pathVariables))
	for index, variable := range pathVariables {
		variables = append(variables, NewPathVariableRow(variable, mapAt(asList(passthrough["variable"]), index)))
	}
	return Url{Text: raw, Query: query, Variable: variables, Raw: without(passthrough, "query", "variable")}
}

func (u Url) MarshalJSON() ([]byte, error) {
	url := newObject().set("raw", u.Text)
	if u.Text != "" {
		parts := NewUrlParts(u.Text)
		url.setIf(parts.Protocol != "", "protocol", parts.Protocol).
			setIf(len(parts.Host) > 0, "host", parts.Host).
			setIf(parts.Port != "", "port", parts.Port).
			setIf(parts.HasPath, "path", parts.Path).
			setIf(parts.Hash != "", "hash", parts.Hash)
	}
	url.setIf(len(u.Query) > 0, "query", u.Query).setIf(len(u.Variable) > 0, "variable", u.Variable)
	return url.merged(without(u.Raw, "protocol", "host", "port", "path", "hash")).MarshalJSON()
}
