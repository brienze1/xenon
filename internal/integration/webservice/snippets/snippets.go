package snippets

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
)

type generator func(request entity.SnippetRequest, given map[string]string) string

var generators = map[enums.SnippetLanguage]generator{
	curlLanguage.Language:           generateCurl,
	jsFetchLanguage.Language:        generateJsFetch,
	goNativeLanguage.Language:       generateGoNative,
	pythonRequestsLanguage.Language: generatePythonRequests,
}

func Languages() []entity.SnippetLanguageSpec {
	return []entity.SnippetLanguageSpec{curlLanguage, jsFetchLanguage, goNativeLanguage, pythonRequestsLanguage}
}

func Generate(request entity.SnippetRequest, options entity.SnippetOptions) (string, bool) {
	generate, found := generators[options.Language]
	if !found {
		return "", false
	}
	return generate(request, options.Options), true
}
