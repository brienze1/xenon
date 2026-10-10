//go:build parity

// Package parity diffs Xenon's script sandbox against postman-sandbox 6.7.4 outcomes; run it with `make parity`.
package parity

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/scriptownerkind"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
	"github.com/brienze1/xenon/internal/integration/webservice/sandbox"
)

const (
	corpusDir        = "../resources/scripts/corpus"
	expectedDir      = corpusDir + "/expected"
	requiredMatchPct = 98.0
)

type keyValue struct {
	Key      string `json:"key"`
	Value    any    `json:"value"`
	Disabled bool   `json:"disabled"`
}

type corpusBody struct {
	Mode       string     `json:"mode"`
	Raw        string     `json:"raw"`
	Urlencoded []keyValue `json:"urlencoded"`
	Formdata   []keyValue `json:"formdata"`
	Graphql    any        `json:"graphql"`
}

type corpusRequest struct {
	Body   *corpusBody `json:"body"`
	URL    any         `json:"url"`
	Method string      `json:"method"`
	Header []keyValue  `json:"header"`
}

type corpusEnvironment struct {
	Name   string     `json:"name"`
	Values []keyValue `json:"values"`
}

type corpusCase struct {
	Vault               map[string]any    `json:"vault"`
	Request             corpusRequest     `json:"request"`
	ID                  string            `json:"id"`
	Phase               string            `json:"phase"`
	Script              string            `json:"script"`
	Environment         corpusEnvironment `json:"environment"`
	CollectionVariables []keyValue        `json:"collectionVariables"`
	Globals             []keyValue        `json:"globals"`
	Variables           []keyValue        `json:"variables"`
}

type outcomeError struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

type outcomeRequest struct {
	Body    any              `json:"body"`
	Method  string           `json:"method"`
	URL     string           `json:"url"`
	Headers []map[string]any `json:"headers"`
}

type outcome struct {
	Error        *outcomeError             `json:"error"`
	Request      *outcomeRequest           `json:"request"`
	Variables    map[string]map[string]any `json:"variables"`
	Vault        map[string]any            `json:"vault"`
	ID           string                    `json:"id"`
	Tests        []any                     `json:"tests"`
	Console      []any                     `json:"console"`
	SentRequests []any                     `json:"sentRequests"`
	Skipped      bool                      `json:"skipped"`
}

func TestSandboxMatchesPostmanOracle(t *testing.T) {
	cases := readCases(t)
	runner := sandbox.NewScriptRunnerWebservice()
	matched := 0
	for _, testCase := range cases {
		t.Run(testCase.ID, func(t *testing.T) {
			expected := readJSON(t, filepath.Join(expectedDir, testCase.ID+".json"))
			actual := normalize(t, runCase(runner, testCase))
			if differences := diff(expected, actual); len(differences) > 0 {
				t.Errorf("differs from postman-sandbox in %s", strings.Join(differences, ", "))
				return
			}
			matched++
		})
	}
	percentage := 100 * float64(matched) / float64(len(cases))
	t.Logf("parity: %d/%d corpus cases match postman-sandbox 6.7.4 (%.1f%%)", matched, len(cases), percentage)
	if percentage < requiredMatchPct {
		t.Errorf("parity %.1f%% is below the required %.1f%% (SC-001)", percentage, requiredMatchPct)
	}
}

func readCases(t *testing.T) []corpusCase {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(corpusDir, "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no corpus cases in %s: %v", corpusDir, err)
	}
	sort.Strings(files)
	cases := make([]corpusCase, 0, len(files))
	for _, file := range files {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var testCase corpusCase
		if err = json.Unmarshal(raw, &testCase); err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		cases = append(cases, testCase)
	}
	return cases
}

func readJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("missing oracle outcome %s (run `make parity`): %v", path, err)
	}
	var decoded map[string]any
	if err = json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}

func normalize(t *testing.T, value outcome) map[string]any {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err = json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}

func diff(expected, actual map[string]any) []string {
	keys := make([]string, 0, len(expected))
	for key := range expected {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	differences := make([]string, 0)
	for _, key := range keys {
		if !reflect.DeepEqual(expected[key], actual[key]) {
			differences = append(differences, key)
		}
	}
	return differences
}

func runCase(runner interface {
	Run(context.Context, entity.ScriptRef, entity.ScriptContext, usecase.SendScriptRequest, entity.AppSettings) entity.ScriptRun
}, testCase corpusCase) outcome {
	phase := scriptphase.PostResponse
	if testCase.Phase == "prerequest" {
		phase = scriptphase.PreRequest
	}
	script := entity.ScriptRef{Phase: phase, OwnerKind: scriptownerkind.Request, ItemID: testCase.ID, OwnerName: testCase.ID, Text: testCase.Script}
	run := runner.Run(context.Background(), script, entity.ScriptContext{Request: entity.NewRequest("", testCase.ID)}, nil, entity.DefaultAppSettings())
	result := outcome{
		ID:           testCase.ID,
		Tests:        []any{},
		Console:      []any{},
		SentRequests: []any{},
		Vault:        valuesOf(testCase.Vault),
		Variables: map[string]map[string]any{
			"environment":         scope(testCase.Environment.Values),
			"collectionVariables": scope(testCase.CollectionVariables),
			"globals":             scope(testCase.Globals),
			"local":               scope(testCase.Variables),
		},
	}
	if run.Error != nil {
		result.Error = errorOf(*run.Error)
	}
	if phase == scriptphase.PreRequest {
		result.Request = requestOf(testCase.Request)
	}
	return result
}

func errorOf(scriptError entity.ScriptError) *outcomeError {
	name, message, found := strings.Cut(scriptError.Message, ": ")
	if !found {
		return &outcomeError{Name: "Error", Message: scriptError.Message}
	}
	return &outcomeError{Name: name, Message: message}
}

func scope(values []keyValue) map[string]any {
	result := map[string]any{}
	for _, value := range values {
		if !value.Disabled {
			result[value.Key] = value.Value
		}
	}
	return result
}

func valuesOf(values map[string]any) map[string]any {
	if values == nil {
		return map[string]any{}
	}
	return values
}

func requestOf(request corpusRequest) *outcomeRequest {
	headers := make([]map[string]any, 0, len(request.Header))
	for _, header := range request.Header {
		if !header.Disabled {
			headers = append(headers, map[string]any{"key": header.Key, "value": header.Value})
		}
	}
	method := request.Method
	if method == "" {
		method = "GET"
	}
	url, _ := request.URL.(string)
	return &outcomeRequest{Method: method, URL: url, Headers: headers, Body: bodyOf(request.Body)}
}

func bodyOf(body *corpusBody) any {
	if body == nil || body.Mode == "" {
		return nil
	}
	switch body.Mode {
	case "raw":
		return map[string]any{"mode": "raw", "raw": body.Raw}
	case "urlencoded":
		return map[string]any{"mode": "urlencoded", "urlencoded": enabled(body.Urlencoded)}
	case "formdata":
		return map[string]any{"mode": "formdata", "formdata": enabled(body.Formdata)}
	default:
		return map[string]any{"mode": body.Mode}
	}
}

func enabled(values []keyValue) []map[string]any {
	result := make([]map[string]any, 0, len(values))
	for _, value := range values {
		if !value.Disabled {
			result = append(result, map[string]any{"key": value.Key, "value": value.Value})
		}
	}
	return result
}
