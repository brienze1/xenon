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
	"sync"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
	"github.com/brienze1/xenon/internal/domain/enums/scripterrorkind"
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
	Value    any    `json:"value"`
	Key      string `json:"key"`
	Type     string `json:"type"`
	Disabled bool   `json:"disabled"`
}

type corpusGraphQl struct {
	Variables     any    `json:"variables"`
	Query         string `json:"query"`
	OperationName string `json:"operationName"`
}

type corpusBody struct {
	Graphql    *corpusGraphQl `json:"graphql"`
	Mode       string         `json:"mode"`
	Raw        string         `json:"raw"`
	Urlencoded []keyValue     `json:"urlencoded"`
	Formdata   []keyValue     `json:"formdata"`
}

type corpusAuth struct {
	Type   string     `json:"type"`
	Bearer []keyValue `json:"bearer"`
	Basic  []keyValue `json:"basic"`
}

type corpusRequest struct {
	Body   *corpusBody `json:"body"`
	Auth   *corpusAuth `json:"auth"`
	URL    string      `json:"url"`
	Method string      `json:"method"`
	Header []keyValue  `json:"header"`
}

type corpusResponse struct {
	Status       string     `json:"status"`
	Body         string     `json:"body"`
	Header       []keyValue `json:"header"`
	Code         int        `json:"code"`
	ResponseTime int64      `json:"responseTime"`
}

type corpusEnvironment struct {
	Name   string     `json:"name"`
	Values []keyValue `json:"values"`
}

type corpusCookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	Secure   bool   `json:"secure"`
	HttpOnly bool   `json:"httpOnly"`
}

type corpusCase struct {
	Vault               map[string]any    `json:"vault"`
	Response            *corpusResponse   `json:"response"`
	Request             corpusRequest     `json:"request"`
	ID                  string            `json:"id"`
	Phase               string            `json:"phase"`
	Script              string            `json:"script"`
	Environment         corpusEnvironment `json:"environment"`
	CollectionVariables []keyValue        `json:"collectionVariables"`
	Globals             []keyValue        `json:"globals"`
	Variables           []keyValue        `json:"variables"`
	Cookies             []corpusCookie    `json:"cookies"`
	SendRequest         []corpusFixture   `json:"sendRequest"`
	RunRequest          []corpusFixture   `json:"runRequest"`
}

type corpusFixture struct {
	corpusResponse
	Error string `json:"error"`
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

type outcomeTest struct {
	Error   *string `json:"error"`
	Name    string  `json:"name"`
	Passed  bool    `json:"passed"`
	Skipped bool    `json:"skipped"`
	Async   bool    `json:"async"`
}

type outcomeLog struct {
	Level string `json:"level"`
	Args  []any  `json:"args"`
}

type outcome struct {
	Error        *outcomeError             `json:"error"`
	Request      *outcomeRequest           `json:"request"`
	Variables    map[string]map[string]any `json:"variables"`
	Vault        map[string]any            `json:"vault"`
	ID           string                    `json:"id"`
	Tests        []outcomeTest             `json:"tests"`
	Console      []outcomeLog              `json:"console"`
	SentRequests []any                     `json:"sentRequests"`
	Skipped      bool                      `json:"skipped"`
}

// fixtureRequests answers pm.sendRequest and pm.execution.runRequest from the case's fixtures, as the oracle does.
type fixtureRequests struct {
	sent  []any
	sends []corpusFixture
	runs  []corpusFixture
	mutex sync.Mutex
}

func (f *fixtureRequests) Send(_ context.Context, request entity.Request, _ entity.ScriptScopes) entity.NestedRequest {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.sent = append(f.sent, requestOutcome(request))
	if len(f.sends) == 0 {
		return entity.NewRejectedNestedRequest("no fixture response")
	}
	fixture := f.sends[0]
	f.sends = f.sends[1:]
	if fixture.Error != "" {
		return entity.NewRejectedNestedRequest(fixture.Error)
	}
	return entity.NewSentNestedRequest(entity.EffectiveRequest{}, responseOf(fixture.corpusResponse))
}

func (f *fixtureRequests) Run(_ context.Context, requestId string, scopes entity.ScriptScopes) entity.NestedRequest {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.sent = append(f.sent, map[string]any{"runRequest": requestId})
	if len(f.runs) == 0 {
		return entity.NewMissingRunNestedRequest(requestId, "no fixture response")
	}
	fixture := f.runs[0]
	f.runs = f.runs[1:]
	response := responseOf(fixture.corpusResponse)
	return entity.NestedRequest{RequestID: requestId, Response: &response, Scopes: scopes}
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
				if os.Getenv("PARITY_VERBOSE") != "" {
					for _, key := range differences {
						want, _ := json.Marshal(expected[key])
						got, _ := json.Marshal(actual[key])
						t.Logf("%s\n  postman: %s\n  xenon:   %s", key, want, got)
					}
				}
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
	requests := &fixtureRequests{sent: []any{}, sends: testCase.SendRequest, runs: testCase.RunRequest}
	run := runner.Run(context.Background(), script, contextOf(testCase), requests, entity.DefaultAppSettings())
	result := outcome{
		ID:           testCase.ID,
		Tests:        []outcomeTest{},
		Console:      []outcomeLog{},
		SentRequests: requests.sent,
		Vault:        map[string]any{},
		Skipped:      run.HasSkippedRequest(),
		Variables: map[string]map[string]any{
			"environment":         scopeOutcome(run.Scopes.Environment),
			"collectionVariables": scopeOutcome(run.Scopes.Collection),
			"globals":             scopeOutcome(run.Scopes.Globals),
			"local":               scopeOutcome(run.Scopes.Local),
		},
	}
	for key, value := range run.Vault {
		result.Vault[key] = value
	}
	for _, log := range run.Logs {
		result.Console = append(result.Console, outcomeLog{Level: log.Level.String(), Args: log.Values})
	}
	for _, test := range run.Tests {
		result.Tests = append(result.Tests, testOutcome(test))
	}
	if run.Error != nil {
		result.Error = errorOf(*run.Error)
	}
	if phase == scriptphase.PreRequest {
		request := requestOutcome(run.Request)
		result.Request = &request
	}
	return result
}

func contextOf(testCase corpusCase) entity.ScriptContext {
	vault := map[string]string{}
	for key, value := range testCase.Vault {
		vault[key] = entity.LocalVariableText(value)
	}
	context := entity.ScriptContext{
		Request: requestOf(testCase.Request),
		Vault:   vault,
		Scopes: entity.ScriptScopes{
			Environment: entity.NewScriptScope(testCase.Environment.Name, scriptValues(testCase.Environment.Values)),
			Collection:  entity.NewScriptScope("", scriptValues(testCase.CollectionVariables)),
			Globals:     entity.NewScriptScope("", scriptValues(testCase.Globals)),
			Local:       entity.NewScriptScope("", scriptValues(testCase.Variables)),
		},
		Cookies: make([]entity.ScriptCookie, 0, len(testCase.Cookies)),
	}
	for _, cookie := range testCase.Cookies {
		context.Cookies = append(context.Cookies, entity.ScriptCookie{Name: cookie.Name, Value: cookie.Value, Domain: cookie.Domain, Path: cookie.Path, Secure: cookie.Secure, HttpOnly: cookie.HttpOnly})
	}
	if testCase.Phase != "prerequest" {
		response := corpusResponse{Code: 200, Status: "OK"}
		if testCase.Response != nil {
			response = *testCase.Response
		}
		built := responseOf(response)
		context.Response = &built
	}
	return context
}

func scriptValues(values []keyValue) []entity.ScriptVariable {
	variables := make([]entity.ScriptVariable, 0, len(values))
	for _, value := range values {
		variables = append(variables, entity.ScriptVariable{Key: value.Key, Value: value.Value, Type: value.Type, Disabled: value.Disabled})
	}
	return variables
}

func requestOf(request corpusRequest) entity.Request {
	built := entity.NewRequest("", "")
	built.Method = request.Method
	built.URL = request.URL
	if built.URL == "" && built.Method == "" {
		built.URL = "https://example.com/"
	}
	if built.Method == "" {
		built.Method = "GET"
	}
	built.Headers = keyValuesOf(request.Header)
	if request.Body != nil {
		built.Body = bodyEntity(*request.Body)
	}
	if request.Auth != nil && request.Auth.Type == "bearer" {
		built.Auth = entity.Auth{Type: authtype.Bearer, Bearer: entity.AuthBearer{Token: parameter(request.Auth.Bearer, "token")}}
	}
	return built
}

func bodyEntity(body corpusBody) entity.Body {
	built := entity.DefaultBody()
	switch body.Mode {
	case "raw":
		built.Mode = bodymode.Raw
		built.Raw = entity.BodyRaw{Language: rawlanguage.Text, Content: body.Raw}
	case "urlencoded":
		built.Mode = bodymode.UrlEncoded
		built.UrlEncoded = keyValuesOf(body.Urlencoded)
	case "formdata":
		built.Mode = bodymode.FormData
		for _, row := range body.Formdata {
			formRow := entity.NewFormDataRow(row.Key, entity.LocalVariableText(row.Value))
			formRow.Enabled = !row.Disabled
			built.FormData = append(built.FormData, formRow)
		}
	case "graphql":
		built.Mode = bodymode.GraphQl
		if body.Graphql != nil {
			built.GraphQl = entity.BodyGraphQl{Query: body.Graphql.Query, Variables: entity.LocalVariableText(body.Graphql.Variables)}
		}
	}
	return built
}

func keyValuesOf(values []keyValue) []entity.KeyValue {
	rows := make([]entity.KeyValue, 0, len(values))
	for _, value := range values {
		rows = append(rows, entity.KeyValue{Key: value.Key, Value: entity.LocalVariableText(value.Value), Enabled: !value.Disabled})
	}
	return rows
}

func parameter(values []keyValue, key string) string {
	for _, value := range values {
		if value.Key == key {
			return entity.LocalVariableText(value.Value)
		}
	}
	return ""
}

func responseOf(fixture corpusResponse) entity.Response {
	response := entity.NewResponse(fixture.Code, fixture.Status, "HTTP/1.1", keyValuesOf(fixture.Header), []byte(fixture.Body), 0, entity.Timings{Total: time.Duration(fixture.ResponseTime) * time.Millisecond})
	return response
}

func scopeOutcome(scope entity.ScriptScope) map[string]any {
	values := map[string]any{}
	for _, variable := range scope.Values {
		if !variable.Disabled {
			values[variable.Key] = variable.Value
		}
	}
	return values
}

func testOutcome(test entity.TestResult) outcomeTest {
	result := outcomeTest{Name: test.Name, Passed: !test.IsFailed(), Skipped: test.IsSkipped(), Async: test.Async}
	if test.IsFailed() {
		message := test.ErrorMessage
		result.Error = &message
	}
	return result
}

func errorOf(scriptError entity.ScriptError) *outcomeError {
	text := scriptError.Message
	if scriptError.Kind == scripterrorkind.UnhandledRejection {
		text = strings.TrimPrefix(text, "Unhandled promise rejection: ")
	}
	name, message, found := strings.Cut(text, ": ")
	if !found {
		return &outcomeError{Name: "Error", Message: scriptError.Message}
	}
	return &outcomeError{Name: name, Message: message}
}

func requestOutcome(request entity.Request) outcomeRequest {
	headers := make([]map[string]any, 0, len(request.Headers))
	for _, header := range request.Headers {
		if header.Enabled {
			headers = append(headers, map[string]any{"key": header.Key, "value": header.Value})
		}
	}
	method := request.Method
	if method == "" {
		method = "GET"
	}
	return outcomeRequest{Method: method, URL: request.URL, Headers: headers, Body: bodyOutcome(request.Body)}
}

func bodyOutcome(body entity.Body) any {
	undefined := map[string]any{"$type": "undefined"}
	switch body.Mode {
	case bodymode.Raw:
		return map[string]any{"mode": "raw", "raw": body.Raw.Content}
	case bodymode.UrlEncoded:
		return map[string]any{"mode": "urlencoded", "urlencoded": enabledRows(body.UrlEncoded)}
	case bodymode.FormData:
		rows := make([]map[string]any, 0, len(body.FormData))
		for _, row := range body.FormData {
			if row.Enabled {
				rows = append(rows, map[string]any{"key": row.Key, "value": row.Value})
			}
		}
		return map[string]any{"mode": "formdata", "formdata": rows}
	case bodymode.GraphQl:
		graphql := map[string]any{"query": body.GraphQl.Query, "operationName": undefined, "variables": undefined}
		if body.GraphQl.Variables != "" {
			graphql["variables"] = body.GraphQl.Variables
		}
		return map[string]any{"mode": "graphql", "graphql": graphql}
	default:
		return nil
	}
}

func enabledRows(values []entity.KeyValue) []map[string]any {
	result := make([]map[string]any, 0, len(values))
	for _, value := range values {
		if value.Enabled {
			result = append(result, map[string]any{"key": value.Key, "value": value.Value})
		}
	}
	return result
}
