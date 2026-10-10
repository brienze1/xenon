package dto

import "go.yaml.in/yaml/v3"

type YamlFormRow struct {
	Name        string    `yaml:"name"`
	Type        string    `yaml:"type"`
	ContentType string    `yaml:"contentType"`
	Value       yaml.Node `yaml:"value"`
	Disabled    bool      `yaml:"disabled"`
}

type YamlFile struct {
	FilePath    string `yaml:"filePath"`
	ContentType string `yaml:"contentType"`
	Selected    bool   `yaml:"selected"`
}

type YamlBody struct {
	Type      string    `yaml:"type"`
	Query     string    `yaml:"query"`
	Variables string    `yaml:"variables"`
	Data      yaml.Node `yaml:"data"`
}

type YamlHttp struct {
	Body    *YamlBody  `yaml:"body"`
	Method  string     `yaml:"method"`
	Url     string     `yaml:"url"`
	Params  []YamlPair `yaml:"params"`
	Headers []YamlPair `yaml:"headers"`
	Auth    yaml.Node  `yaml:"auth"`
}

type YamlRuntime struct {
	Variables  []YamlPair      `yaml:"variables"`
	Assertions []YamlAssertion `yaml:"assertions"`
	Scripts    []YamlScript    `yaml:"scripts"`
}

type YamlRequest struct {
	Http     YamlHttp    `yaml:"http"`
	Docs     string      `yaml:"docs"`
	Info     YamlInfo    `yaml:"info"`
	Runtime  YamlRuntime `yaml:"runtime"`
	Settings yaml.Node   `yaml:"settings"`
}

func (y *YamlRequest) ToRequest(file string) Request {
	request := Request{
		Name:     y.Info.Name,
		Type:     y.Info.Type,
		Seq:      y.Info.Seq,
		Tags:     y.Info.Tags,
		File:     file,
		Method:   y.Http.Method,
		Url:      y.Http.Url,
		Params:   []Pair{},
		Path:     []Pair{},
		Headers:  yamlPairs(y.Http.Headers),
		Settings: yamlSettings(y.Settings),
		Auth:     yamlAuth(y.Http.Auth),
		Body:     y.Http.Body.toBody(),
		Docs:     trimFinalNewline(y.Docs),
		Scripts:  yamlScripts(y.Runtime.Scripts, y.Runtime.Variables, y.Runtime.Assertions),
	}
	for _, param := range y.Http.Params {
		if param.Type == pathParam {
			request.Path = append(request.Path, param.ToPair())
			continue
		}
		request.Params = append(request.Params, param.ToPair())
	}
	return request
}

func (y *YamlBody) toBody() Body {
	body := Body{Pairs: []Pair{}, Multipart: []MultipartRow{}, Files: []FilePart{}}
	if y == nil {
		return body
	}
	body.Mode = y.Type
	if mode, found := yamlBodyModes[y.Type]; found {
		body.Mode = mode
	}
	switch body.Mode {
	case FormUrlEncodedBody:
		var rows []YamlPair
		_ = y.Data.Decode(&rows)
		body.Pairs = yamlPairs(rows)
	case MultipartFormBody:
		var rows []YamlFormRow
		_ = y.Data.Decode(&rows)
		for _, row := range rows {
			body.Multipart = append(body.Multipart, row.toMultipart())
		}
	case FileBody:
		var files []YamlFile
		_ = y.Data.Decode(&files)
		for _, file := range files {
			body.Files = append(body.Files, FilePart{Path: file.FilePath, ContentType: file.ContentType, Selected: file.Selected})
		}
	case GraphQlBody:
		body.Query = trimFinalNewline(y.Query)
		body.Variables = trimFinalNewline(y.Variables)
	default:
		body.Text = trimFinalNewline(y.Data.Value)
	}
	return body
}

func (y YamlFormRow) toMultipart() MultipartRow {
	row := MultipartRow{Key: y.Name, ContentType: y.ContentType, Enabled: !y.Disabled, Files: []string{}}
	if y.Type != FileBody {
		row.Value = y.Value.Value
		return row
	}
	if y.Value.Kind == yaml.SequenceNode {
		_ = y.Value.Decode(&row.Files)
	} else {
		row.Files = append(row.Files, y.Value.Value)
	}
	return row
}
