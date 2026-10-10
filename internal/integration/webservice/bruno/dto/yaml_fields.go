package dto

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	beforeRequestScript = "before-request"
	afterResponseScript = "after-response"
	testsScript         = "tests"
	preRequestScope     = "pre-request"
	queryParam          = "query"
	pathParam           = "path"
	authTypeKey         = "type"
)

var yamlBodyModes = map[string]string{"form-urlencoded": FormUrlEncodedBody, "multipart-form": MultipartFormBody}

type YamlInfo struct {
	Seq  *int     `yaml:"seq"`
	Name string   `yaml:"name"`
	Type string   `yaml:"type"`
	Tags []string `yaml:"tags"`
}

type YamlPair struct {
	Enabled  *bool  `yaml:"enabled"`
	Name     string `yaml:"name"`
	Value    string `yaml:"value"`
	Type     string `yaml:"type"`
	Scope    string `yaml:"scope"`
	Line     int    `yaml:"-"`
	Disabled bool   `yaml:"disabled"`
	Secret   bool   `yaml:"secret"`
}

func (y *YamlPair) UnmarshalYAML(node *yaml.Node) error {
	type plain YamlPair
	if err := node.Decode((*plain)(y)); err != nil {
		return err
	}
	y.Line = node.Line
	return nil
}

func (y YamlPair) IsEnabled() bool {
	return !y.Disabled && (y.Enabled == nil || *y.Enabled)
}

func (y YamlPair) ToPair() Pair {
	return Pair{Key: y.Name, Value: y.Value, Line: y.Line, Enabled: y.IsEnabled()}
}

type YamlScript struct {
	Type string `yaml:"type"`
	Code string `yaml:"code"`
	Line int    `yaml:"-"`
}

func (y *YamlScript) UnmarshalYAML(node *yaml.Node) error {
	type plain YamlScript
	if err := node.Decode((*plain)(y)); err != nil {
		return err
	}
	for index := 0; index+1 < len(node.Content); index += 2 {
		if code := node.Content[index+1]; node.Content[index].Value == "code" {
			y.Line = code.Line
			if code.Style&(yaml.LiteralStyle|yaml.FoldedStyle) != 0 {
				y.Line++
			}
		}
	}
	return nil
}

type YamlAssertion struct {
	Enabled    *bool  `yaml:"enabled"`
	Expression string `yaml:"expression"`
	Operator   string `yaml:"operator"`
	Value      string `yaml:"value"`
	Line       int    `yaml:"-"`
	Disabled   bool   `yaml:"disabled"`
}

func (y *YamlAssertion) UnmarshalYAML(node *yaml.Node) error {
	type plain YamlAssertion
	if err := node.Decode((*plain)(y)); err != nil {
		return err
	}
	y.Line = node.Line
	return nil
}

type YamlRequestSettings struct {
	Headers   []YamlPair      `yaml:"headers"`
	Variables []YamlPair      `yaml:"variables"`
	Scripts   []YamlScript    `yaml:"scripts"`
	Asserts   []YamlAssertion `yaml:"assertions"`
	Auth      yaml.Node       `yaml:"auth"`
}

func yamlScripts(scripts []YamlScript, variables []YamlPair, assertions []YamlAssertion) Scripts {
	result := Scripts{PreRequestVars: []Pair{}, PostResponseVars: []Pair{}, Assertions: []entity.BrunoAssertion{}}
	for _, script := range scripts {
		code := entity.BrunoScript{Text: trimFinalNewline(script.Code), Line: script.Line}
		switch script.Type {
		case beforeRequestScript:
			result.PreRequest = code
		case afterResponseScript:
			result.PostResponse = code
		case testsScript:
			result.Tests = code
		}
	}
	for _, variable := range variables {
		if variable.Scope == "" || variable.Scope == preRequestScope {
			result.PreRequestVars = append(result.PreRequestVars, variable.ToPair())
			continue
		}
		result.PostResponseVars = append(result.PostResponseVars, variable.ToPair())
	}
	for _, assertion := range assertions {
		enabled := !assertion.Disabled && (assertion.Enabled == nil || *assertion.Enabled)
		result.Assertions = append(result.Assertions, entity.BrunoAssertion{Expression: assertion.Expression, Operator: assertion.Operator, Value: assertion.Value, Line: assertion.Line, Enabled: enabled})
	}
	return result
}

func yamlAuth(node yaml.Node) Auth {
	auth := Auth{Values: []Pair{}}
	for index := 0; index+1 < len(node.Content); index += 2 {
		key, value := node.Content[index].Value, node.Content[index+1].Value
		if key == authTypeKey {
			auth.Mode = value
			continue
		}
		auth.Values = append(auth.Values, NewPair(key, value, true))
	}
	return auth
}

func yamlPairs(rows []YamlPair) []Pair {
	result := make([]Pair, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.ToPair())
	}
	return result
}

func yamlSettings(node yaml.Node) []Pair {
	result := make([]Pair, 0)
	for index := 0; index+1 < len(node.Content); index += 2 {
		result = append(result, NewPair(node.Content[index].Value, node.Content[index+1].Value, true))
	}
	return result
}

func trimFinalNewline(text string) string {
	for len(text) > 0 && text[len(text)-1] == '\n' {
		text = text[:len(text)-1]
	}
	return text
}
