package snippets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
	"github.com/brienze1/xenon/internal/domain/enums/snippetlanguage"
)

var corpus = filepath.Join("..", "..", "..", "..", "test", "integration", "resources", "snippets")

type corpusParam struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Src      string `json:"src"`
	Type     string `json:"type"`
	Disabled bool   `json:"disabled"`
}

type corpusRequest struct {
	Method string        `json:"method"`
	URL    string        `json:"url"`
	Header []corpusParam `json:"header"`
	Body   *struct {
		Mode       string        `json:"mode"`
		Raw        string        `json:"raw"`
		UrlEncoded []corpusParam `json:"urlencoded"`
		FormData   []corpusParam `json:"formdata"`
		File       corpusParam   `json:"file"`
		GraphQl    struct {
			Query     string `json:"query"`
			Variables string `json:"variables"`
		} `json:"graphql"`
	} `json:"body"`
}

func (c corpusRequest) snippetRequest() entity.SnippetRequest {
	request := entity.SnippetRequest{Method: c.Method, URL: c.URL, Body: entity.DefaultBody()}
	for _, header := range c.Header {
		request.Headers = append(request.Headers, entity.NewKeyValue(header.Key, header.Value))
	}
	if c.Body == nil {
		return request
	}
	switch c.Body.Mode {
	case "raw":
		request.Body.Mode, request.Body.Raw = bodymode.Raw, entity.BodyRaw{Language: rawlanguage.Text, Content: c.Body.Raw}
	case "urlencoded":
		request.Body.Mode = bodymode.UrlEncoded
		for _, row := range c.Body.UrlEncoded {
			request.Body.UrlEncoded = append(request.Body.UrlEncoded, entity.KeyValue{Key: row.Key, Value: row.Value, Enabled: !row.Disabled})
		}
	case "formdata":
		request.Body.Mode = bodymode.FormData
		for _, row := range c.Body.FormData {
			kind := formdatatype.Text
			if row.Type == "file" {
				kind = formdatatype.File
			}
			request.Body.FormData = append(request.Body.FormData, entity.FormDataRow{Key: row.Key, Value: row.Value, Src: entity.FileRef{Src: row.Src}, Type: kind, Enabled: !row.Disabled})
		}
	case "graphql":
		request.Body.Mode, request.Body.GraphQl = bodymode.GraphQl, entity.BodyGraphQl{Query: c.Body.GraphQl.Query, Variables: c.Body.GraphQl.Variables}
		request.Headers = append(request.Headers, entity.NewKeyValue(entity.HeaderContentType, jsonContentType))
	case "file":
		request.Body.Mode, request.Body.Binary = bodymode.Binary, entity.FileRef{Src: c.Body.File.Src}
	}
	return request
}

func corpusOptions(t *testing.T) map[string]map[string]map[string]any {
	content, err := os.ReadFile(filepath.Join(corpus, "options.json"))
	if err != nil {
		t.Fatal(err)
	}
	var sets map[string]map[string]map[string]any
	if err := json.Unmarshal(content, &sets); err != nil {
		t.Fatal(err)
	}
	return sets
}

func optionStrings(values map[string]any) map[string]string {
	result := map[string]string{}
	for key, value := range values {
		result[key] = fmt.Sprint(value)
	}
	return result
}

func TestGeneratorsMatchTheGoldenCorpus(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(corpus, "requests", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no corpus requests: %v", err)
	}
	sets := corpusOptions(t)
	compared := 0
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var parsed corpusRequest
		if err := json.Unmarshal(content, &parsed); err != nil {
			t.Fatal(err)
		}
		name := strings.TrimSuffix(filepath.Base(file), ".json")
		for language, languageSets := range sets {
			names := make([]string, 0, len(languageSets))
			for set := range languageSets {
				names = append(names, set)
			}
			sort.Strings(names)
			for _, set := range names {
				expected, err := os.ReadFile(filepath.Join(corpus, "expected", name+"."+language+"."+set+".txt"))
				if err != nil {
					t.Fatal(err)
				}
				text, found := Generate(parsed.snippetRequest(), entity.SnippetOptions{Language: snippetlanguage.GetSnippetLanguage(language), Options: optionStrings(languageSets[set])})
				compared++
				if !found || text != string(expected) {
					t.Errorf("%s.%s.%s differs:\n--- expected ---\n%s\n--- actual ---\n%s", name, language, set, expected, text)
				}
			}
		}
	}
	if compared != 130 {
		t.Errorf("compared %d snippets, want 130", compared)
	}
}
