package dto

import (
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	httpRequestKind    = "http-request"
	graphqlRequestKind = "graphql-request"
	requestKindSuffix  = "-request"
	graphqlMethod      = "POST"
	extensionsKey      = "postmanV3"
)

type Request struct {
	Extra         map[string]any `yaml:",inline"`
	Settings      map[string]any `yaml:"settings"`
	Body          *Body          `yaml:"body"`
	Order         *float64       `yaml:"order"`
	ID            any            `yaml:"id"`
	Description   any            `yaml:"description"`
	Examples      yaml.Node      `yaml:"examples"`
	Variables     yaml.Node      `yaml:"variables"`
	Kind          string         `yaml:"$kind"`
	Name          string         `yaml:"name"`
	URL           string         `yaml:"url"`
	Method        string         `yaml:"method"`
	Query         string         `yaml:"query"`
	Headers       KeyValues      `yaml:"headers"`
	QueryParams   KeyValues      `yaml:"queryParams"`
	PathVariables KeyValues      `yaml:"pathVariables"`
	Scripts       []Script       `yaml:"scripts"`
	Auth          Auth           `yaml:"auth"`
	examples      []Example
}

func IsSupportedRequestKind(kind string) bool {
	return kind == httpRequestKind || kind == graphqlRequestKind
}

func IsRequestKind(kind string) bool {
	return strings.HasSuffix(kind, requestKindSuffix)
}

func (r *Request) inlineExamples() ([]Example, error) {
	if r.Examples.Kind != yaml.SequenceNode {
		return []Example{}, nil
	}
	var examples []Example
	return examples, r.Examples.Decode(&examples)
}

func (r *Request) examplesPath() string {
	if r.Examples.Kind == yaml.ScalarNode {
		return r.Examples.Value
	}
	return ""
}

func (r *Request) toPostmanV21(conversion *Conversion, parentPath string) (map[string]any, error) {
	path := entity.JoinImportPath(parentPath, r.Name)
	extensions := map[string]any{}
	item := copied(r.Extra)
	item["name"] = r.Name
	setIfPresent(item, "id", r.ID)
	setIfPresent(item, "description", r.Description)
	request, err := r.requestToPostmanV21(conversion, path, extensions)
	if err != nil {
		return nil, err
	}
	item["request"] = request
	setIfPresent(item, "event", scriptsToPostmanV21(r.Scripts, conversion, r.Name, path, extensions))
	setIfPresent(item, "protocolProfileBehavior", settingsToPostmanV21(r.Settings))
	if r.Kind == httpRequestKind && r.Variables.Kind == yaml.SequenceNode {
		var variables KeyValues
		if err := r.Variables.Decode(&variables); err != nil {
			return nil, err
		}
		item["variable"] = variables.toPostmanV21()
	}
	responses := make([]any, 0, len(r.examples))
	for index := range r.examples {
		response, err := r.examples[index].toPostmanV21()
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	setIfPresent(item, "response", responses)
	setIfPresent(item, extensionsKey, extensions)
	return item, nil
}

func (r *Request) requestToPostmanV21(conversion *Conversion, path string, extensions map[string]any) (map[string]any, error) {
	request := map[string]any{"method": r.Method, "header": r.Headers.toPostmanV21(), "url": urlToPostmanV21(r.URL, r.QueryParams, r.PathVariables)}
	body, err := r.Body.toPostmanV21()
	if err != nil {
		return nil, err
	}
	if r.Kind == graphqlRequestKind {
		request["method"] = graphqlMethod
		body = map[string]any{"mode": graphqlBodyType, graphqlBodyType: map[string]any{"query": r.Query, "variables": r.graphqlVariables()}}
	}
	setIfPresent(request, "body", body)
	setIfPresent(request, "auth", r.Auth.toPostmanV21(conversion, r.Name, path, extensions))
	return request, nil
}

func (r *Request) graphqlVariables() string {
	var variables any
	if err := r.Variables.Decode(&variables); err != nil {
		return ""
	}
	return text(variables)
}

func settingsToPostmanV21(settings map[string]any) map[string]any {
	behavior := map[string]any{}
	for key, value := range settings {
		setIfPresent(behavior, key, value)
	}
	return behavior
}
