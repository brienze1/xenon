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
	pathParam           = "path"
	authTypeKey         = "type"
	openCollectionKey   = "opencollection"
)

var yamlBodyModes = map[string]string{"form-urlencoded": FormUrlEncodedBody, "multipart-form": MultipartFormBody}

func NewRequestFromYaml(source YamlRequest, file string) Request {
	request := Request{
		Name:     nameOrFileName(source.Info.Name, file),
		Type:     source.Info.Type,
		Seq:      source.Info.Seq,
		Tags:     source.Info.Tags,
		File:     file,
		Method:   source.Http.Method,
		Url:      source.Http.Url,
		Params:   []Pair{},
		Path:     []Pair{},
		Headers:  yamlPairs(source.Http.Headers),
		Settings: yamlSettings(source.Settings),
		Auth:     yamlAuth(source.Http.Auth),
		Body:     yamlBody(source.Http.Body),
		Docs:     trimFinalNewline(source.Docs),
		Scripts:  yamlScripts(source.Runtime.Scripts, source.Runtime.Variables, source.Runtime.Assertions),
	}
	for _, param := range source.Http.Params {
		if param.Type == pathParam {
			request.Path = append(request.Path, param.ToPair())
			continue
		}
		request.Params = append(request.Params, param.ToPair())
	}
	return request
}

func NewFolderFromYaml(source YamlFolder, file, fallbackName string) Folder {
	folder := Folder{
		Name:    source.Info.Name,
		Seq:     source.Info.Seq,
		File:    file,
		Docs:    trimFinalNewline(source.Docs),
		Headers: yamlPairs(source.Request.Headers),
		Auth:    yamlAuth(source.Request.Auth),
		Scripts: yamlScripts(source.Request.Scripts, source.Request.Variables, source.Request.Asserts),
	}
	if folder.Name == "" {
		folder.Name = fallbackName
	}
	return folder
}

func NewCollectionFromYaml(source YamlCollection, file string) Collection {
	scripts := yamlScripts(source.Request.Scripts, source.Request.Variables, source.Request.Asserts)
	collection := Collection{
		Name:      source.Info.Name,
		File:      file,
		Docs:      trimFinalNewline(source.Docs),
		Headers:   yamlPairs(source.Request.Headers),
		Auth:      yamlAuth(source.Request.Auth),
		Variables: scripts.PreRequestVars,
		Scripts:   scripts,
		Config:    map[string]any{openCollectionKey: source.OpenCollection},
	}
	collection.Scripts.PreRequestVars = []Pair{}
	return collection
}

func NewEnvironmentFromYaml(source YamlEnvironment, fallbackName string) Environment {
	environment := NewEnvironment(source.Name, source.Color)
	if environment.Name == "" {
		environment.Name = fallbackName
	}
	for _, variable := range source.Variables {
		environment.Variables = append(environment.Variables, NewEnvironmentVariable(variable.Name, variable.Value, variable.IsEnabled(), variable.Secret))
	}
	return environment
}

func yamlBody(source *YamlBody) Body {
	if source == nil {
		return NewBody("")
	}
	mode := source.Type
	if mapped, found := yamlBodyModes[source.Type]; found {
		mode = mapped
	}
	body := NewBody(mode)
	switch body.Mode {
	case FormUrlEncodedBody:
		var rows []YamlPair
		_ = source.Data.Decode(&rows)
		body.Pairs = yamlPairs(rows)
	case MultipartFormBody:
		var rows []YamlFormRow
		_ = source.Data.Decode(&rows)
		for _, row := range rows {
			body.Multipart = append(body.Multipart, yamlMultipartRow(row))
		}
	case FileBody:
		var files []YamlFile
		_ = source.Data.Decode(&files)
		for _, file := range files {
			body.Files = append(body.Files, NewFilePart(file.FilePath, file.ContentType, file.Selected))
		}
	case GraphQlBody:
		body.Query = trimFinalNewline(source.Query)
		body.Variables = trimFinalNewline(source.Variables)
	default:
		body.Text = trimFinalNewline(source.Data.Value)
	}
	return body
}

func yamlMultipartRow(source YamlFormRow) MultipartRow {
	row := MultipartRow{Key: source.Name, ContentType: source.ContentType, Enabled: !source.Disabled, Files: []string{}}
	if source.Type != FileBody {
		row.Value = source.Value.Value
		return row
	}
	if source.Value.Kind == yaml.SequenceNode {
		_ = source.Value.Decode(&row.Files)
	} else {
		row.Files = append(row.Files, source.Value.Value)
	}
	return row
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
		result.Assertions = append(result.Assertions, assertion.ToEntity())
	}
	return result
}

func yamlAuth(node yaml.Node) Auth {
	auth := NewAuth("", []Pair{})
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
