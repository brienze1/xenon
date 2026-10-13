package snippets

import (
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
	"github.com/brienze1/xenon/internal/domain/enums/snippetlanguage"
)

func TestJavascriptNumbersAndStringsFollowJsonStringify(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{input: `[0, -0, 1.0, 1E3, 1e21, 1.5e21, 0.000001, 1e-7, -2.5e-8, 123456789012345680000, 1e400, 0.1]`, want: `[0,0,1,1000,1e+21,1.5e+21,0.000001,1e-7,-2.5e-8,123456789012345680000,null,0.1]`},
		{input: `{"b":1,"2":true,"1":false,"a":null,"10":[],"x":{}}`, want: `{"1":false,"2":true,"10":[],"b":1,"a":null,"x":{}}`},
		{input: `{"a":1,"a":2}`, want: `{"a":2}`},
		{input: "\"tab\\t\\u0001\\u001f<&>\u00e9\\\"\\\\\"", want: "\"tab\\t\\u0001\\u001f<&>\u00e9\\\"\\\\\""},
		{input: `{"4294967295":1,"01":2,"3":3}`, want: `{"3":3,"4294967295":1,"01":2}`},
	}
	for _, tc := range cases {
		parsed, valid := parseJson(tc.input)
		if !valid {
			t.Fatalf("%s did not parse", tc.input)
		}
		if got := stringifyJson(parsed, 0, javascriptLiterals); got != tc.want {
			t.Errorf("stringify(%s) = %s, want %s", tc.input, got, tc.want)
		}
	}
	for _, invalid := range []string{"", "{", "[1,]", "1 2", `{"a":}`, "nul"} {
		if _, valid := parseJson(invalid); valid {
			t.Errorf("%q should not parse", invalid)
		}
	}
	parsed, _ := parseJson(`{"a":[true,null],"b":{}}`)
	if got := stringifyJson(parsed, 12, pythonLiterals); got != "{\n          \"a\": [\n                    True,\n                    None\n          ],\n          \"b\": {}\n}" {
		t.Errorf("python indent: %q", got)
	}
}

func TestCurlOptionsAndEdgeCases(t *testing.T) {
	request := entity.SnippetRequest{
		Method:  "HEAD",
		URL:     "https://x.test/{{id}}?a=1&flag&&b=[x]#frag",
		Headers: []entity.KeyValue{entity.NewKeyValue("X-Empty", ""), entity.NewKeyValue("", "skipped"), entity.NewKeyValue("X-Quote", `say "hi" \"there\" $HOME!`)},
	}
	got := generateCurl(request, map[string]string{optionSilent: "true", optionQuoteType: "double", optionTimeoutSeconds: "2.5", optionMultiLine: "false"})
	want := `curl --silent --location --max-time 2.5 --globoff --head "https://x.test/{{id}}?a=1&flag&b=[x]\#frag" --header "X-Empty;" --header "X-Quote: say \"hi\" \\\"there\\\" \$HOME\!"`
	if got != want {
		t.Errorf("curl =\n%s\nwant\n%s", got, want)
	}
	cases := []struct {
		request entity.SnippetRequest
		options map[string]string
		want    string
	}{
		{request: entity.SnippetRequest{Method: "GET", URL: "https://x", Body: entity.Body{Mode: bodymode.Raw, Raw: entity.BodyRaw{Content: "x"}}}, want: "curl --location --request GET 'https://x' \\\n--data 'x'"},
		{request: entity.SnippetRequest{Method: "POST", URL: "https://x"}, want: "curl --location --request POST 'https://x'"},
		{request: entity.SnippetRequest{Method: "POST", URL: "https://x", Body: entity.Body{Mode: bodymode.Raw, Raw: entity.BodyRaw{Content: `{"a": 1}`}}}, options: map[string]string{optionMultiLine: "false", optionLongFormat: "false"}, want: `curl -L 'https://x' -d '{"a":1}'`},
		{request: entity.SnippetRequest{Method: "PATCH", URL: "https://x", Body: entity.Body{Mode: bodymode.Binary, Binary: entity.FileRef{Src: " a.bin "}}}, options: map[string]string{optionLongFormat: "false", optionTrimBody: "true", optionFollowRedirect: "false"}, want: "curl -X PATCH 'https://x' \\\n-d '@a.bin'"},
		{request: entity.SnippetRequest{Method: "GET", URL: "https://x"}, options: map[string]string{optionFollowMethod: "true"}, want: "curl --location --request GET 'https://x'"},
		{request: entity.SnippetRequest{Method: "POST", URL: "https://x", Body: entity.Body{Mode: bodymode.FormData, FormData: []entity.FormDataRow{{Key: "off", Value: "x"}, {Key: "on", Value: `a\b`, Enabled: true}}}}, want: "curl --location 'https://x' \\\n--form 'on=\"a\\\\b\"'"},
	}
	for index, tc := range cases {
		if got := generateCurl(tc.request, tc.options); got != tc.want {
			t.Errorf("case %d:\n%s\nwant\n%s", index, got, tc.want)
		}
	}
}

func TestOtherGeneratorsEdgeCases(t *testing.T) {
	request := entity.SnippetRequest{Method: "POST", URL: "https://x", Notices: []string{"Note one."}, Body: entity.Body{Mode: bodymode.UrlEncoded, UrlEncoded: []entity.KeyValue{{Key: "off", Value: "x"}}}}
	if got := generatePythonRequests(request, nil); !strings.HasPrefix(got, "# Note one.\nimport requests\n") || !strings.Contains(got, "payload = {}\n") {
		t.Errorf("python:\n%s", got)
	}
	request.Body = entity.Body{Mode: bodymode.FormData, FormData: []entity.FormDataRow{{Key: "off", Type: formdatatype.Text}}}
	if got := generatePythonRequests(request, nil); !strings.Contains(got, "payload = {}\nfiles={}\n") || !strings.Contains(got, "files=files") {
		t.Errorf("python formdata:\n%s", got)
	}
	request.Body = entity.Body{Mode: bodymode.FormData, FormData: []entity.FormDataRow{{Key: "doc", Type: formdatatype.File, Src: entity.FileRef{Src: "dir/report.weird"}, Enabled: true}}}
	if got := generatePythonRequests(request, map[string]string{optionTimeout: "3"}); !strings.Contains(got, "('doc',('report.weird',open('dir/report.weird','rb'),'application/octet-stream'))") || !strings.Contains(got, "timeout=3") {
		t.Errorf("python file:\n%s", got)
	}
	jsRequest := entity.SnippetRequest{Method: "GET", URL: "https://x", Headers: []entity.KeyValue{entity.NewKeyValue("Content-Type", "application/vnd.api+json")}, Body: entity.Body{Mode: bodymode.Raw, Raw: entity.BodyRaw{Content: `{"a":1}`}}}
	got := generateJsFetch(jsRequest, map[string]string{optionAsyncAwait: "true", optionTimeout: "100", optionIndentType: "Tab", optionIndentCount: "1"})
	for _, want := range []string{"JSON.stringify({\n \"a\": 1\n})", "try {\n\tconst response = await fetch(\"https://x\", requestOptions);", "} finally {\n\tclearTimeout(timerId);\n};"} {
		if !strings.Contains(got, want) {
			t.Errorf("js fetch lacks %q:\n%s", want, got)
		}
	}
	if got := generateJsFetch(entity.SnippetRequest{Method: "GET", URL: "https://x"}, map[string]string{optionAsyncAwait: "true"}); !strings.HasSuffix(got, "console.error(error);\n};") {
		t.Errorf("js fetch async:\n%s", got)
	}
	goRequest := entity.SnippetRequest{Method: "PUT", URL: "https://x", Body: entity.Body{Mode: bodymode.Raw, Raw: entity.BodyRaw{Content: "a`b\r\n"}}}
	if got := generateGoNative(goRequest, map[string]string{optionTimeout: "1500"}); !strings.Contains(got, "strings.NewReader(`a`+\"`\"+`b`+\"\r\"+`\n`)") || !strings.Contains(got, "time.Duration(1.5 * time.Second)") {
		t.Errorf("go:\n%s", got)
	}
	if _, found := Generate(goRequest, entity.SnippetOptions{Language: snippetlanguage.Unknown}); found {
		t.Error("an unknown language should not generate")
	}
	if len(Languages()) != 4 {
		t.Errorf("languages: %v", Languages())
	}
}
