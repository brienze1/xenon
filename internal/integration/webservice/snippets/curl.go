package snippets

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
	"github.com/brienze1/xenon/internal/domain/enums/snippetlanguage"
)

const (
	singleQuote = "'"
	doubleQuote = `"`
)

var curlLongForms = map[string]string{
	"-s": "--silent",
	"-L": "--location",
	"-m": "--max-time",
	"-I": "--head",
	"-X": "--request",
	"-H": "--header",
	"-d": "--data",
	"-F": "--form",
	"-g": "--globoff",
}

var curlDoubleQuoteEscapes = strings.NewReplacer("`", "\\`", "#", `\#`, "$", `\$`, "!", `\!`)

var curlLanguage = entity.SnippetLanguageSpec{
	Language: snippetlanguage.Curl,
	Label:    "cURL",
	Options: []entity.SnippetOptionSpec{
		{ID: optionMultiLine, Label: "Generate multiline snippet", Type: booleanOption, Default: "true"},
		{ID: optionLongFormat, Label: "Use long form options", Type: booleanOption, Default: "true"},
		{ID: optionLineContinuation, Label: "Line continuation character", Type: enumOption, Default: `\`, Values: []string{`\`, "^", "`"}},
		{ID: optionQuoteType, Label: "Quote type", Type: enumOption, Default: "single", Values: []string{"single", "double"}},
		{ID: optionTimeoutSeconds, Label: "Set request timeout (in seconds)", Type: positiveOption, Default: "0"},
		followRedirectOption,
		{ID: optionFollowMethod, Label: "Follow original HTTP method", Type: booleanOption, Default: "false"},
		trimBodyOption,
		{ID: optionSilent, Label: "Use silent mode", Type: booleanOption, Default: "false"},
	},
}

type curlWriter struct {
	options options
	quote   string
	indent  string
	request entity.SnippetRequest
	long    bool
	trim    bool
}

func generateCurl(request entity.SnippetRequest, given map[string]string) string {
	resolved := resolveOptions(curlLanguage.Options, given)
	writer := curlWriter{request: request, options: resolved, quote: singleQuote, indent: " ", long: resolved.flag(optionLongFormat), trim: resolved.flag(optionTrimBody)}
	if resolved.text(optionQuoteType) == "double" {
		writer.quote = doubleQuote
	}
	if resolved.flag(optionMultiLine) {
		writer.indent = " " + resolved.text(optionLineContinuation) + "\n"
	}
	return commentLines(request.Notices, "#") + writer.command()
}

func (c curlWriter) form(option string) string {
	if c.long {
		return curlLongForms[option]
	}
	return option
}

func (c curlWriter) sanitize(text string, trim bool, quote string, backSlash, urlEncode bool) string {
	if urlEncode {
		text = encodeURIComponent(text)
	}
	if backSlash {
		text = strings.ReplaceAll(text, `\`, `\\`)
	}
	switch quote {
	case doubleQuote:
		text = curlDoubleQuoteEscapes.Replace(escapeEscapedQuotes(strings.ReplaceAll(text, `"`, `\"`)))
	case singleQuote:
		text = strings.ReplaceAll(text, "'", `'\''`)
	}
	return trimIf(text, trim)
}

func escapeEscapedQuotes(text string) string {
	var builder strings.Builder
	for index := 0; index < len(text); index++ {
		if strings.HasPrefix(text[index:], `\\"`) && (index == 0 || text[index-1] != '\\') {
			builder.WriteString(`\\\"`)
			index += 2
			continue
		}
		builder.WriteByte(text[index])
	}
	return builder.String()
}

func (c curlWriter) command() string {
	url := c.sanitize(encodedQueryUrl(c.request.URL), false, c.quote, false, false)
	snippet := "curl"
	if c.options.flag(optionSilent) {
		snippet += " " + c.form("-s")
	}
	if c.options.flag(optionFollowRedirect) {
		snippet += " " + c.form("-L")
	}
	if timeout := c.options.number(optionTimeoutSeconds); timeout > 0 {
		snippet += " " + c.form("-m") + " " + javascriptNumber(timeout)
	}
	if strings.ContainsAny(url, "{[}]") {
		snippet += " " + c.form("-g")
	}
	if c.request.Method == "HEAD" {
		snippet += " " + c.form("-I")
	}
	if c.addsMethod() {
		snippet += " " + c.form("-X") + " " + c.request.Method
	}
	snippet += " " + c.quote + url + c.quote
	return snippet + c.headers() + c.body()
}

func (c curlWriter) addsMethod() bool {
	if c.options.flag(optionFollowRedirect) && c.options.flag(optionFollowMethod) {
		return true
	}
	switch c.request.Method {
	case "HEAD":
		return false
	case "GET":
		return !c.request.IsBodyEmpty()
	case "POST":
		return c.request.IsBodyEmpty()
	default:
		return true
	}
}

func (c curlWriter) headers() string {
	snippet := ""
	for _, header := range c.request.Headers {
		if header.Key == "" {
			continue
		}
		snippet += c.indent + c.form("-H") + " " + c.quote + c.sanitize(header.Key, true, c.quote, false, false)
		if header.Value != "" {
			snippet += ": " + c.sanitize(header.Value, false, c.quote, false, false) + c.quote
			continue
		}
		snippet += ";" + c.quote
	}
	return snippet
}

func (c curlWriter) body() string {
	if !c.request.HasBody() {
		return ""
	}
	body := c.request.Body
	switch body.Mode {
	case bodymode.UrlEncoded:
		return c.urlEncoded(body.UrlEncoded)
	case bodymode.Raw:
		return c.raw(body.Raw.Content)
	case bodymode.GraphQl:
		return c.data(graphQlPayload(body.GraphQl))
	case bodymode.FormData:
		return c.formData(body.FormData)
	default:
		option := "-d"
		if c.long {
			option = "--data-binary"
		}
		return c.indent + option + " " + c.quote + "@" + trimIf(body.Binary.Src, c.trim) + c.quote
	}
}

func (c curlWriter) urlEncoded(rows []entity.KeyValue) string {
	option := "-d"
	if c.long {
		option = "--data-urlencode"
	}
	snippet := ""
	for _, row := range rows {
		if row.Enabled {
			snippet += c.indent + option + " " + c.quote + c.sanitize(row.Key, c.trim, c.quote, false, true) + "=" + c.sanitize(row.Value, c.trim, c.quote, false, !c.long) + c.quote
		}
	}
	return snippet
}

func (c curlWriter) raw(content string) string {
	sanitized := c.sanitize(content, c.trim, c.quote, false, false)
	if !c.options.flag(optionMultiLine) {
		if parsed, valid := parseJson(sanitized); valid {
			sanitized = stringifyJson(parsed, 0, javascriptLiterals)
		}
	}
	return c.indent + c.dataOption(content) + " " + c.quote + sanitized + c.quote
}

func (c curlWriter) data(payload string) string {
	return c.indent + c.dataOption(payload) + " " + c.quote + c.sanitize(payload, c.trim, c.quote, false, false) + c.quote
}

func (c curlWriter) dataOption(payload string) string {
	if strings.Contains(payload, "@") {
		return "--data-raw"
	}
	return c.form("-d")
}

func (c curlWriter) formData(rows []entity.FormDataRow) string {
	snippet := ""
	for _, row := range rows {
		if !row.Enabled {
			continue
		}
		snippet += c.indent + c.form("-F") + " " + c.quote + c.sanitize(row.Key, c.trim, c.quote, false, false) + "="
		if row.Type == formdatatype.File {
			snippet += c.sanitize(`@"`+c.sanitize(row.Src.Src, c.trim, doubleQuote, true, false)+`"`, c.trim, c.quote, c.quote == doubleQuote, false) + c.quote
			continue
		}
		snippet += c.sanitize(`"`+c.sanitize(row.Value, c.trim, doubleQuote, true, false)+`"`, c.trim, c.quote, c.quote == doubleQuote, false) + c.quote
	}
	return snippet
}

func graphQlPayload(graphQl entity.BodyGraphQl) string {
	return stringifyJson(jsonValue{kind: jsonObject, members: []jsonMember{
		{key: "query", value: jsonValue{kind: jsonString, text: graphQl.Query}},
		{key: "variables", value: graphQlVariables(graphQl)},
	}}, 0, javascriptLiterals)
}

func graphQlVariables(graphQl entity.BodyGraphQl) jsonValue {
	if variables, valid := parseJson(graphQl.Variables); valid {
		return variables
	}
	return jsonValue{kind: jsonObject}
}
