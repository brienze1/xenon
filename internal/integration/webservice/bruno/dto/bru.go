package dto

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	metaBlock          = "meta"
	httpBlock          = "http"
	authBlock          = "auth"
	authBlockPrefix    = "auth:"
	bodyBlockPrefix    = "body:"
	docsBlock          = "docs"
	settingsBlock      = "settings"
	headersBlock       = "headers"
	queryBlock         = "params:query"
	pathBlock          = "params:path"
	preVarsBlock       = "vars:pre-request"
	postVarsBlock      = "vars:post-response"
	assertBlock        = "assert"
	preScriptBlock     = "script:pre-request"
	postScriptBlock    = "script:post-response"
	testsBlock         = "tests"
	graphqlVarsBlock   = "body:graphql:vars"
	formBlock          = "body:form-urlencoded"
	multipartBlock     = "body:multipart-form"
	fileBlock          = "body:file"
	environmentVars    = "vars"
	environmentSecrets = "vars:secret"
	colorAttribute     = "color"
)

var (
	methodBlocks      = []string{"get", "post", "put", "delete", "patch", "options", "head", "connect", "trace"}
	fileAnnotation    = regexp.MustCompile(`@file\(([^)]*)\)`)
	contentAnnotation = regexp.MustCompile(`@contentType\(([^)]*)\)`)
	unsupportedBlocks = map[string]string{"grpc": "grpc", "ws": "ws"}
)

func NewRequestFromBru(file BruFile) Request {
	request := Request{
		Name:     file.Value(metaBlock, "name"),
		Type:     file.Value(metaBlock, "type"),
		Seq:      sequence(file.Value(metaBlock, "seq")),
		File:     file.Path,
		Tags:     tags(file),
		Params:   pairs(file.Entries(queryBlock)),
		Path:     pairs(file.Entries(pathBlock)),
		Headers:  pairs(file.Entries(headersBlock)),
		Settings: pairs(file.Entries(settingsBlock)),
		Scripts:  bruScripts(file),
	}
	method, block := methodBlock(file)
	request.Method = method
	request.Url = block.Value("url")
	request.Auth = bruAuth(file, block.Value("auth"))
	request.Body = bruBody(file, block.Value("body"))
	request.Docs = textOf(file, docsBlock)
	if request.Type == "" {
		request.Type = unsupportedType(file)
	}
	return request
}

func NewFolderFromBru(file BruFile, fallbackName string) Folder {
	folder := Folder{
		Name:    file.Value(metaBlock, "name"),
		Seq:     sequence(file.Value(metaBlock, "seq")),
		File:    file.Path,
		Headers: pairs(file.Entries(headersBlock)),
		Auth:    bruAuth(file, file.Value(authBlock, "mode")),
		Docs:    textOf(file, docsBlock),
		Scripts: bruScripts(file),
	}
	if folder.Name == "" {
		folder.Name = fallbackName
	}
	return folder
}

func NewCollectionFromBru(file BruFile, name string, config map[string]any) Collection {
	scripts := bruScripts(file)
	collection := Collection{
		Name:      name,
		Config:    config,
		File:      file.Path,
		Headers:   pairs(file.Entries(headersBlock)),
		Auth:      bruAuth(file, file.Value(authBlock, "mode")),
		Docs:      textOf(file, docsBlock),
		Variables: scripts.PreRequestVars,
		Scripts:   scripts,
	}
	collection.Scripts.PreRequestVars = []Pair{}
	return collection
}

func NewEnvironmentFromBru(file BruFile, name string) Environment {
	environment := Environment{Name: name, Color: file.Attribute(colorAttribute), Variables: []EnvironmentVariable{}}
	for _, entry := range file.Entries(environmentVars) {
		environment.Variables = append(environment.Variables, EnvironmentVariable{Key: entry.Key, Value: entry.Value, Enabled: entry.Enabled})
	}
	for _, entry := range file.Entries(environmentSecrets) {
		environment.Variables = append(environment.Variables, EnvironmentVariable{Key: entry.Key, Enabled: entry.Enabled, Secret: true})
	}
	return environment
}

func methodBlock(file BruFile) (string, BruBlock) {
	for _, block := range file.Blocks {
		if contains(methodBlocks, block.Name) {
			return block.Name, block
		}
		if block.Name == httpBlock {
			return block.Value("method"), block
		}
	}
	return "", BruBlock{}
}

func unsupportedType(file BruFile) string {
	for _, block := range file.Blocks {
		if kind, found := unsupportedBlocks[block.Name]; found {
			return kind
		}
	}
	return ""
}

func bruAuth(file BruFile, mode string) Auth {
	return Auth{Mode: mode, Values: pairs(file.Entries(authBlockPrefix + mode))}
}

func bruBody(file BruFile, mode string) Body {
	body := Body{Mode: mode, Pairs: []Pair{}, Multipart: []MultipartRow{}, Files: []FilePart{}}
	switch mode {
	case FormUrlEncodedBody:
		body.Pairs = pairs(file.Entries(formBlock))
	case MultipartFormBody:
		for _, entry := range file.Entries(multipartBlock) {
			body.Multipart = append(body.Multipart, multipartRow(entry))
		}
	case FileBody:
		for _, entry := range file.Entries(fileBlock) {
			body.Files = append(body.Files, FilePart{Path: annotation(fileAnnotation, entry.Value), ContentType: annotation(contentAnnotation, entry.Value), Selected: entry.Enabled})
		}
	case GraphQlBody:
		body.Query = textOf(file, bodyBlockPrefix+GraphQlBody)
		body.Variables = textOf(file, graphqlVarsBlock)
	default:
		body.Text = textOf(file, bodyBlockPrefix+mode)
	}
	return body
}

func multipartRow(entry BruEntry) MultipartRow {
	row := MultipartRow{Key: entry.Key, Value: entry.Value, Enabled: entry.Enabled, Files: []string{}, ContentType: annotation(contentAnnotation, entry.Value)}
	if files := annotation(fileAnnotation, entry.Value); fileAnnotation.MatchString(entry.Value) {
		row.Files = strings.Split(files, "|")
		row.Value = ""
	}
	return row
}

func bruScripts(file BruFile) Scripts {
	return Scripts{
		PreRequest:       script(file, preScriptBlock),
		PostResponse:     script(file, postScriptBlock),
		Tests:            script(file, testsBlock),
		PreRequestVars:   pairs(file.Entries(preVarsBlock)),
		PostResponseVars: pairs(file.Entries(postVarsBlock)),
		Assertions:       assertions(file.Entries(assertBlock)),
	}
}

func script(file BruFile, name string) entity.BrunoScript {
	block, found := file.Block(name)
	if !found {
		return entity.BrunoScript{}
	}
	return entity.BrunoScript{Text: block.Text, Line: block.TextLine}
}

func assertions(entries []BruEntry) []entity.BrunoAssertion {
	result := make([]entity.BrunoAssertion, 0, len(entries))
	for _, entry := range entries {
		operator, value, _ := strings.Cut(strings.TrimSpace(entry.Value), " ")
		result = append(result, entity.BrunoAssertion{Expression: entry.Key, Operator: operator, Value: strings.TrimSpace(value), Line: entry.Line, Enabled: entry.Enabled})
	}
	return result
}

func pairs(entries []BruEntry) []Pair {
	result := make([]Pair, 0, len(entries))
	for _, entry := range entries {
		result = append(result, Pair{Key: entry.Key, Value: entry.Value, Line: entry.Line, Enabled: entry.Enabled})
	}
	return result
}

func tags(file BruFile) []string {
	block, _ := file.Block(metaBlock)
	entry, _ := block.Lookup("tags")
	return entry.Values
}

func textOf(file BruFile, name string) string {
	block, _ := file.Block(name)
	return block.Text
}

func annotation(pattern *regexp.Regexp, value string) string {
	if match := pattern.FindStringSubmatch(value); match != nil {
		return match[1]
	}
	return ""
}

func sequence(value string) *int {
	if number, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
		return &number
	}
	return nil
}
