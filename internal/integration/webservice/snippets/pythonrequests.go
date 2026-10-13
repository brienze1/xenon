package snippets

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
	"github.com/brienze1/xenon/internal/domain/enums/snippetlanguage"
)

const octetStream = "application/octet-stream"

var pythonRequestsLanguage = entity.SnippetLanguageSpec{
	Language: snippetlanguage.PythonRequests,
	Label:    "Python - Requests",
	Options:  []entity.SnippetOptionSpec{indentCountOption, indentTypeOption, timeoutOption, followRedirectOption, trimBodyOption},
}

var pythonQuoteEscapes = strings.NewReplacer(`\`, `\\`, "'", `\'`)

var pythonUrlEscapes = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

var pythonFileTypes = map[string]string{
	"aac": "audio/aac", "abw": "application/x-abiword", "arc": "application/x-freearc", "avi": "video/x-msvideo",
	"azw": "application/vnd.amazon.ebook", "bin": octetStream, "bmp": "image/bmp", "bz": "application/x-bzip",
	"bz2": "application/x-bzip2", "csh": "application/x-csh", "css": "text/css", "csv": "text/csv",
	"doc": "application/msword", "docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"eot": "application/vnd.ms-fontobject", "epub": "application/epub+zip", "gif": "image/gif", "htm": "text/html",
	"html": "text/html", "ico": "image/vnd.microsoft.icon", "ics": "text/calendar", "jar": "application/java-archive",
	"jpeg": "image/jpeg", "jpg": "image/jpeg", "js": "text/javascript", "json": "application/json",
	"jsonld": "application/ld+json", "mid": "audio/midi", "midi": "audio/midi", "mjs": "text/javascript",
	"mp3": "audio/mpeg", "mpeg": "video/mpeg", "mpkg": "application/vnd.apple.installer+xml",
	"odp": "application/vnd.oasis.opendocument.presentation", "ods": "application/vnd.oasis.opendocument.spreadsheet",
	"odt": "application/vnd.oasis.opendocument.text", "oga": "audio/ogg", "ogv": "video/ogg", "ogx": "application/ogg",
	"otf": "font/otf", "png": "image/png", "pdf": "application/pdf", "ppt": "application/vnd.ms-powerpoint",
	"pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation", "rar": "application/x-rar-compressed",
	"rtf": "application/rtf", "sh": "application/x-sh", "svg": "image/svg+xml", "swf": "application/x-shockwave-flash",
	"tar": "application/x-tar", "tif": "image/tiff", "tiff": "image/tiff", "ts": "video/mp2t", "ttf": "font/ttf",
	"txt": "text/plain", "vsd": "application/vnd.visio", "wav": "audio/wav", "weba": "audio/webm", "webm": "video/webm",
	"webp": "image/webp", "woff": "font/woff", "woff2": "font/woff2", "xhtml": "application/xhtml+xml",
	"xls": "application/vnd.ms-excel", "xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"xml": "text/xml", "xul": "application/vnd.mozilla.xul+xml", "zip": "application/zip", "3gp": "video/3gpp",
	"7z": "application/x-7z-compressed", "7-zip": "application/x-7z-compressed",
}

func generatePythonRequests(request entity.SnippetRequest, given map[string]string) string {
	resolved := resolveOptions(pythonRequestsLanguage.Options, given)
	indent := resolved.indent()
	contentType, _ := request.Header(entity.HeaderContentType)
	snippet := "import requests\n"
	if isJsonContentType(contentType) {
		snippet += "import json\n"
	}
	snippet += "\nurl = \"" + pythonUrlEscapes.Replace(request.URL) + "\"\n\n"
	snippet += pythonBody(request, indent, resolved.flag(optionTrimBody), contentType)
	snippet += pythonHeaders(request.Headers, indent) + "\n"
	snippet += `response = requests.request("` + request.Method + `", url, headers=headers`
	if request.HasBody() && request.Body.Mode == bodymode.FormData {
		snippet += ", data=payload, files=files"
	} else {
		snippet += ", data=payload"
	}
	if !resolved.flag(optionFollowRedirect) {
		snippet += ", allow_redirects=False"
	}
	if timeout := resolved.number(optionTimeout); timeout != 0 {
		snippet += ", timeout=" + javascriptNumber(timeout)
	}
	snippet += ")\n\nprint(response.text)\n"
	return commentLines(request.Notices, "#") + snippet
}

func pythonHeaders(headers []entity.KeyValue, indent string) string {
	if len(headers) == 0 {
		return "headers = {}\n"
	}
	lines := make([]string, 0, len(headers))
	for _, header := range headers {
		lines = append(lines, indent+"'"+pythonQuoteEscapes.Replace(strings.TrimSpace(header.Key))+"': '"+pythonQuoteEscapes.Replace(header.Value)+"'")
	}
	return "headers = {\n" + strings.Join(lines, ",\n") + "\n}\n"
}

func pythonString(text string, trim bool) string {
	return quoteJson(trimIf(text, trim))
}

func pythonBody(request entity.SnippetRequest, indent string, trim bool, contentType string) string {
	if !request.HasBody() {
		return "payload = {}\n"
	}
	body := request.Body
	switch body.Mode {
	case bodymode.Raw:
		if parsed, valid := parseJson(body.Raw.Content); valid && isJsonContentType(contentType) {
			return "payload = json.dumps(" + stringifyJson(parsed, len(indent), pythonLiterals) + ")\n"
		}
		return "payload = " + pythonString(body.Raw.Content, trim) + "\n"
	case bodymode.GraphQl:
		return "payload = " + pythonString(graphQlPayload(body.GraphQl), trim) + "\n"
	case bodymode.UrlEncoded:
		pairs := make([]string, 0, len(body.UrlEncoded))
		for _, row := range body.UrlEncoded {
			if row.Enabled {
				pairs = append(pairs, pythonUrlEncoded(row.Key, trim)+"="+pythonUrlEncoded(row.Value, trim))
			}
		}
		if len(pairs) == 0 {
			return "payload = {}\n"
		}
		return "payload = '" + strings.Join(pairs, "&") + "'\n"
	case bodymode.FormData:
		return pythonFormData(body.FormData, indent, trim)
	default:
		return "payload = \"<file contents here>\"\n"
	}
}

func pythonUrlEncoded(text string, trim bool) string {
	return strings.ReplaceAll(encodeURIComponent(trimIf(text, trim)), "'", `\'`)
}

func pythonFormData(rows []entity.FormDataRow, indent string, trim bool) string {
	fields := make([]string, 0)
	files := make([]string, 0)
	for _, row := range rows {
		if !row.Enabled {
			continue
		}
		if row.Type == formdatatype.File {
			source := row.Src.Src
			fileType, found := pythonFileTypes[lastSegment(source, ".")]
			if !found {
				fileType = octetStream
			}
			files = append(files, indent+"('"+row.Key+"',('"+lastSegment(source, "/")+"',open('"+pythonQuoteEscapes.Replace(trimIf(source, trim))+"','rb'),'"+fileType+"'))")
			continue
		}
		fields = append(fields, "'"+pythonQuoteEscapes.Replace(trimIf(row.Key, trim))+"': '"+pythonQuoteEscapes.Replace(trimIf(row.Value, trim))+"'")
	}
	if len(fields) == 0 && len(files) == 0 {
		return "payload = {}\nfiles={}\n"
	}
	return "payload = {" + strings.Join(fields, ",\n") + "}\nfiles=[\n" + strings.Join(files, ",\n") + "\n]\n"
}
