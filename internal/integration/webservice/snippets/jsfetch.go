package snippets

import (
	"regexp"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
	"github.com/brienze1/xenon/internal/domain/enums/snippetlanguage"
)

const jsonContentType = "application/json"

var jsonSuffix = regexp.MustCompile(`\+json$`)

var jsFetchLanguage = entity.SnippetLanguageSpec{
	Language: snippetlanguage.JsFetch,
	Label:    "JavaScript - Fetch",
	Options: []entity.SnippetOptionSpec{
		indentCountOption,
		indentTypeOption,
		timeoutOption,
		followRedirectOption,
		trimBodyOption,
		{ID: optionAsyncAwait, Label: "Use async/await", Type: booleanOption, Default: "false"},
	},
}

var jsFetchBodyNames = map[string]string{
	bodymode.UrlEncoded.String(): "urlencoded",
	bodymode.Raw.String():        "raw",
	bodymode.GraphQl.String():    "graphql",
	bodymode.FormData.String():   "formdata",
	bodymode.Binary.String():     "file",
}

func isJsonContentType(contentType string) bool {
	return contentType == jsonContentType || jsonSuffix.MatchString(contentType)
}

func generateJsFetch(request entity.SnippetRequest, given map[string]string) string {
	resolved := resolveOptions(jsFetchLanguage.Options, given)
	indent := resolved.indent()
	trim := resolved.flag(optionTrimBody)
	timeout := resolved.number(optionTimeout)
	headers := jsFetchHeaders(request.Headers)
	contentType, _ := request.Header(entity.HeaderContentType)
	body := jsFetchBody(request, trim, indent, contentType)
	code := ""
	if timeout > 0 {
		code += "const controller = new AbortController();\n"
		code += "const timerId = setTimeout(() => controller.abort(), " + javascriptNumber(timeout) + ");\n"
	}
	requestOptions := "const requestOptions = {\n" + indent + `method: "` + request.Method + "\",\n" + indent
	if headers != "" {
		requestOptions += "headers: myHeaders,\n" + indent
		code += headers + "\n"
	}
	if body != "" {
		requestOptions += "body: " + jsFetchBodyNames[request.Body.Mode.String()] + ",\n" + indent
		code += body + "\n"
	}
	if timeout > 0 {
		requestOptions += "signal: controller.signal,\n" + indent
	}
	requestOptions += `redirect: "` + redirectMode(resolved.flag(optionFollowRedirect)) + "\"\n};\n"
	code += requestOptions + "\n"
	return commentLines(request.Notices, "//") + code + jsFetchCall(escapeString(request.URL, false), indent, timeout > 0, resolved.flag(optionAsyncAwait))
}

func redirectMode(follow bool) string {
	if follow {
		return "follow"
	}
	return "manual"
}

func jsFetchHeaders(headers []entity.KeyValue) string {
	if len(headers) == 0 {
		return ""
	}
	snippet := "const myHeaders = new Headers();\n"
	for _, header := range headers {
		snippet += `myHeaders.append("` + escapeString(header.Key, true) + `", "` + escapeString(header.Value, false) + "\");\n"
	}
	return snippet
}

func jsFetchBody(request entity.SnippetRequest, trim bool, indent, contentType string) string {
	if !request.HasBody() {
		return ""
	}
	body := request.Body
	switch body.Mode {
	case bodymode.UrlEncoded:
		snippet := "const urlencoded = new URLSearchParams();\n"
		for _, row := range body.UrlEncoded {
			if row.Enabled {
				snippet += `urlencoded.append("` + escapeString(row.Key, trim) + `", "` + escapeString(row.Value, trim) + "\");\n"
			}
		}
		return snippet
	case bodymode.Raw:
		if parsed, valid := parseJson(body.Raw.Content); valid && isJsonContentType(contentType) {
			return "const raw = JSON.stringify(" + stringifyJson(parsed, len(indent), javascriptLiterals) + ");\n"
		}
		return `const raw = "` + escapeString(body.Raw.Content, trim) + "\";\n"
	case bodymode.GraphQl:
		return "const graphql = JSON.stringify({\n" + indent + `query: "` + escapeString(body.GraphQl.Query, trim) + "\",\n" + indent + "variables: " + stringifyJson(graphQlVariables(body.GraphQl), 0, javascriptLiterals) + "\n})"
	case bodymode.FormData:
		snippet := "const formdata = new FormData();\n"
		for _, row := range body.FormData {
			if !row.Enabled {
				continue
			}
			if row.Type == formdatatype.File {
				snippet += `formdata.append("` + escapeString(row.Key, trim) + `", fileInput.files[0], "` + lastSegment(row.Src.Src, "/") + "\");\n"
				continue
			}
			snippet += `formdata.append("` + escapeString(row.Key, trim) + `", "` + escapeString(row.Value, trim) + "\");\n"
		}
		return snippet
	default:
		return "const file = \"<file contents here>\";\n"
	}
}

func jsFetchCall(url, indent string, timeout, asyncAwait bool) string {
	if asyncAwait {
		call := "try {\n" + indent + `const response = await fetch("` + url + "\", requestOptions);\n" + indent
		call += "const result = await response.text();\n" + indent + "console.log(result)\n"
		call += "} catch (error) {\n" + indent + "console.error(error);\n"
		if timeout {
			call += "} finally {\n" + indent + "clearTimeout(timerId);\n"
		}
		return call + "};"
	}
	call := `fetch("` + url + "\", requestOptions)\n" + indent
	call += ".then((response) => response.text())\n" + indent
	call += ".then((result) => console.log(result))\n" + indent
	call += ".catch((error) => console.error(error))"
	if timeout {
		call += "\n" + indent + ".finally(() => clearTimeout(timerId))"
	}
	return call + ";"
}
