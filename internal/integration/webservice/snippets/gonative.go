package snippets

import (
	"strconv"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
	"github.com/brienze1/xenon/internal/domain/enums/snippetlanguage"
)

const millisecondsPerSecond = 1000

var goNativeLanguage = entity.SnippetLanguageSpec{
	Language: snippetlanguage.GoNative,
	Label:    "Go - Native",
	Options:  []entity.SnippetOptionSpec{indentCountOption, indentTypeOption, timeoutOption, followRedirectOption, trimBodyOption},
}

var goRawStringEscapes = strings.NewReplacer("`", "`+\"`\"+`", "\r", "`+\"\r\"+`")

func generateGoNative(request entity.SnippetRequest, given map[string]string) string {
	resolved := resolveOptions(goNativeLanguage.Options, given)
	indent := resolved.indent()
	timeout := resolved.number(optionTimeout)
	body, opensFiles := goNativeBody(request, resolved.flag(optionTrimBody), indent)
	formData := request.HasBody() && request.Body.Mode == bodymode.FormData
	code := "package main\n\nimport (\n" + indent + "\"fmt\"\n"
	if timeout > 0 {
		code += indent + "\"time\"\n"
	}
	switch {
	case formData:
		code += indent + "\"bytes\"\n" + indent + "\"mime/multipart\"\n"
	case body != "":
		code += indent + "\"strings\"\n"
	}
	if opensFiles {
		code += indent + "\"os\"\n" + indent + "\"path/filepath\"\n"
	}
	code += indent + "\"net/http\"\n" + indent + "\"io\"\n)\n\n"
	code += "func main() {\n\n" + indent + `url := "` + escapeString(encodedQueryUrl(request.URL), false) + "\"\n"
	code += indent + `method := "` + request.Method + "\"\n\n"
	if body != "" {
		code += indent + body + "\n\n"
	}
	if timeout > 0 {
		code += indent + "timeout := time.Duration(" + javascriptNumber(timeout/millisecondsPerSecond) + " * time.Second)\n"
	}
	code += indent + "client := &http.Client {\n"
	if !resolved.flag(optionFollowRedirect) {
		code += strings.Repeat(indent, 2) + "CheckRedirect: func(req *http.Request, via []*http.Request) error {\n"
		code += strings.Repeat(indent, 3) + "return http.ErrUseLastResponse\n" + strings.Repeat(indent, 2) + "},\n"
	}
	if timeout > 0 {
		code += strings.Repeat(indent, 2) + "Timeout: timeout,\n"
	}
	code += indent + "}\n"
	payload := "nil"
	if body != "" {
		payload = "payload"
	}
	code += indent + "req, err := http.NewRequest(method, url, " + payload + ")\n\n"
	code += goNativeFailure(indent, "err")
	if headers := goNativeHeaders(request.Headers, indent); headers != "" {
		code += headers + "\n"
	}
	if formData {
		code += indent + "req.Header.Set(\"Content-Type\", writer.FormDataContentType())\n"
	}
	code += indent + "res, err := client.Do(req)\n" + goNativeFailure(indent, "err")
	code += indent + "defer res.Body.Close()\n\n" + indent + "body, err := io.ReadAll(res.Body)\n" + goNativeFailure(indent, "err")
	code += indent + "fmt.Println(string(body))\n}"
	return commentLines(request.Notices, "//") + code
}

func goNativeFailure(indent, variable string) string {
	return indent + "if " + variable + " != nil {\n" + strings.Repeat(indent, 2) + "fmt.Println(" + variable + ")\n" + strings.Repeat(indent, 2) + "return\n" + indent + "}\n"
}

func goNativeHeaders(headers []entity.KeyValue, indent string) string {
	snippet := ""
	for _, header := range headers {
		snippet += indent + `req.Header.Add("` + escapeString(header.Key, true) + `", "` + escapeString(header.Value, false) + "\")\n"
	}
	return snippet
}

func goNativeBody(request entity.SnippetRequest, trim bool, indent string) (string, bool) {
	if !request.HasBody() {
		return "", false
	}
	body := request.Body
	switch body.Mode {
	case bodymode.UrlEncoded:
		pairs := make([]string, 0, len(body.UrlEncoded))
		for _, row := range body.UrlEncoded {
			if row.Enabled {
				pairs = append(pairs, encodeURIComponent(row.Key)+"="+encodeURIComponent(row.Value))
			}
		}
		return `payload := strings.NewReader("` + strings.Join(pairs, "&") + "\")", false
	case bodymode.Raw:
		return "payload := strings.NewReader(`" + trimIf(goRawStringEscapes.Replace(body.Raw.Content), trim) + "`)", false
	case bodymode.GraphQl:
		return `payload := strings.NewReader("` + escapeString(graphQlPayload(body.GraphQl), trim) + "\")", false
	case bodymode.FormData:
		return goNativeFormData(body.FormData, trim, indent)
	default:
		return "payload := strings.NewReader(\"<file contents here>\")\n", false
	}
}

func goNativeFormData(rows []entity.FormDataRow, trim bool, indent string) (string, bool) {
	snippet := "payload := &bytes.Buffer{}\n" + indent + "writer := multipart.NewWriter(payload)\n"
	opensFiles := false
	for index, row := range rows {
		if !row.Enabled {
			continue
		}
		position := strconv.Itoa(index + 1)
		if row.Type != formdatatype.File {
			snippet += indent + `_ = writer.WriteField("` + escapeString(row.Key, trim) + `", "` + escapeString(row.Value, trim) + "\")\n"
			continue
		}
		opensFiles = true
		snippet += indent + "file, errFile" + position + " := os.Open(\"" + row.Src.Src + "\")\n"
		snippet += indent + "defer file.Close()\n"
		snippet += indent + "part" + position + ",\n         errFile" + position + " := writer.CreateFormFile(\"" + escapeString(row.Key, trim) + "\",filepath.Base(\"" + row.Src.Src + "\"))\n"
		snippet += indent + "_, errFile" + position + " = io.Copy(part" + position + ", file)\n"
		snippet += indent + "if errFile" + position + " != nil {\n" + strings.Repeat(indent, 2) + "fmt.Println(errFile" + position + ")\n" + strings.Repeat(indent, 2) + "return\n" + indent + "}\n"
	}
	snippet += indent + "err := writer.Close()\n" + goNativeFailure(indent, "err")
	return snippet, opensFiles
}
