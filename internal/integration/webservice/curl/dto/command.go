package dto

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
	"github.com/brienze1/xenon/internal/domain/enums/httpmethod"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
)

const (
	filePrefix         = "@"
	jsonContentType    = "application/json"
	cookieHeader       = "Cookie"
	refererHeader      = "Referer"
	usernameAttribute  = "username"
	passwordAttribute  = "password"
	formUrlEncodedType = "application/x-www-form-urlencoded"
)

type Command struct {
	User       *string
	Method     string
	UploadFile string
	Urls       []string
	Headers    []entity.KeyValue
	Data       []DataPart
	Json       []string
	Forms      []string
	Notices    []string
	Get        bool
	Head       bool
	Digest     bool
	Insecure   bool
	Location   bool
}

func NewCommand() *Command {
	return &Command{Urls: []string{}, Headers: []entity.KeyValue{}, Data: []DataPart{}, Json: []string{}, Forms: []string{}, Notices: []string{}}
}

func (c *Command) AddHeader(raw string) {
	key, value, _ := strings.Cut(raw, ":")
	c.Headers = append(c.Headers, entity.NewKeyValue(strings.TrimSpace(key), strings.TrimSpace(value)))
}

func (c *Command) SetHeader(key, value string) {
	c.Headers = append(c.Headers, entity.NewKeyValue(key, value))
}

func (c *Command) AddData(value string) {
	c.Data = append(c.Data, NewDataPart(value, false))
}

func (c *Command) AddRawData(value string) {
	c.Data = append(c.Data, NewDataPart(value, true))
}

func (c *Command) AddUrlEncodedData(value string) {
	c.Data = append(c.Data, NewUrlEncodedPart(value))
}

func (c *Command) AddForm(form string) {
	c.Forms = append(c.Forms, form)
}

func (c *Command) AddJson(json string) {
	c.Json = append(c.Json, json)
}

func (c *Command) AddUrl(address string) {
	c.Urls = append(c.Urls, address)
}

func (c *Command) SetUser(user string) {
	c.User = &user
}

func (c *Command) SetUserAgent(userAgent string) {
	c.SetHeader(entity.HeaderUserAgent, userAgent)
}

func (c *Command) SetReferer(referer string) {
	c.SetHeader(refererHeader, referer)
}

func (c *Command) AddCookie(cookie string) {
	if !strings.Contains(cookie, "=") {
		c.Notice("Ignored the cookie file " + cookie)
		return
	}
	for index, header := range c.Headers {
		if header.HasKey(cookieHeader) {
			c.Headers[index].Value += "; " + cookie
			return
		}
	}
	c.SetHeader(cookieHeader, cookie)
}

func (c *Command) Notice(notice string) {
	c.Notices = append(c.Notices, notice)
}

func (c *Command) IgnoreUrlsOtherThan(address string) {
	for _, other := range c.Urls {
		if other != address {
			c.Notice("Ignored the extra URL " + other)
		}
	}
}

func (c *Command) ToEntity(address string) entity.ParsedCurl {
	request := entity.NewRequest("", "")
	request.URL = address
	if c.Get {
		request.URL = withQuery(address, c.queryParts())
	} else {
		request.Body = c.body()
	}
	request.Headers = c.headers()
	request.Method = c.method(request.Body)
	request.Auth = c.auth()
	request.Settings.VerifySsl = !c.Insecure
	request.Settings.FollowRedirects = request.Settings.FollowRedirects || c.Location
	if request.Body.Mode == bodymode.Raw {
		request.Body.Raw.Language = rawLanguage(c.contentType(), request.Body.Raw.Content)
	}
	return entity.ParsedCurl{Request: request, Notices: c.Notices}
}

func (c *Command) method(body entity.Body) string {
	switch {
	case c.Method != "":
		return strings.ToUpper(c.Method)
	case c.Head:
		return httpmethod.Head.String()
	case c.UploadFile != "":
		return httpmethod.Put.String()
	case body.HasContent():
		return httpmethod.Post.String()
	default:
		return httpmethod.Get.String()
	}
}

func (c *Command) headers() []entity.KeyValue {
	headers := append([]entity.KeyValue{}, c.Headers...)
	if len(c.Json) == 0 {
		return headers
	}
	for _, name := range []string{entity.HeaderContentType, entity.HeaderAccept} {
		if !hasHeader(headers, name) {
			headers = append(headers, entity.NewKeyValue(name, jsonContentType))
		}
	}
	return headers
}

func (c *Command) auth() entity.Auth {
	if c.User == nil {
		return entity.DefaultAuth()
	}
	username, password, _ := strings.Cut(*c.User, ":")
	if c.Digest {
		return entity.Auth{Type: authtype.Digest, Attributes: []entity.AuthAttribute{{Key: usernameAttribute, Value: username}, {Key: passwordAttribute, Value: password}}}
	}
	return entity.Auth{Type: authtype.Basic, Basic: entity.AuthBasic{Username: username, Password: password}}
}

func (c *Command) body() entity.Body {
	body := entity.DefaultBody()
	switch {
	case len(c.Json) > 0:
		body.Mode = bodymode.Raw
		body.Raw = entity.NewBodyRaw(rawlanguage.Json.String(), strings.Join(c.Json, ""))
	case len(c.Forms) > 0:
		body.Mode = bodymode.FormData
		for _, form := range c.Forms {
			body.FormData = append(body.FormData, formRow(form))
		}
	case c.UploadFile != "":
		body.Mode = bodymode.Binary
		body.Binary = entity.NewRepickFileRef(c.UploadFile)
	case c.dataFile() != "":
		body.Mode = bodymode.Binary
		body.Binary = entity.NewRepickFileRef(c.dataFile())
	case len(c.Data) > 0 && c.isFormEncoded():
		body.Mode = bodymode.UrlEncoded
		body.UrlEncoded = c.urlEncodedRows()
	case len(c.Data) > 0:
		body.Mode = bodymode.Raw
		body.Raw = entity.NewBodyRaw(rawlanguage.Text.String(), c.joinedData())
	}
	return body
}

func (c *Command) dataFile() string {
	for _, part := range c.Data {
		if part.IsFile() {
			return strings.TrimPrefix(part.Value, filePrefix)
		}
	}
	return ""
}

func (c *Command) isFormEncoded() bool {
	if contentType := c.contentType(); contentType != "" {
		return strings.HasPrefix(strings.ToLower(contentType), formUrlEncodedType)
	}
	formLike := true
	for _, part := range c.Data {
		if part.UrlEncoded {
			return true
		}
		formLike = formLike && part.IsFormLike()
	}
	return formLike
}

func (c *Command) urlEncodedRows() []entity.KeyValue {
	rows := make([]entity.KeyValue, 0)
	for _, part := range c.Data {
		rows = append(rows, part.Rows()...)
	}
	return rows
}

func (c *Command) queryParts() []string {
	parts := make([]string, 0, len(c.Data))
	for _, part := range c.Data {
		parts = append(parts, part.Query())
	}
	return parts
}

func (c *Command) joinedData() string {
	parts := make([]string, 0, len(c.Data))
	for _, part := range c.Data {
		parts = append(parts, part.Value)
	}
	return strings.Join(parts, "&")
}

func (c *Command) contentType() string {
	for _, header := range c.Headers {
		if header.HasKey(entity.HeaderContentType) {
			return header.Value
		}
	}
	return ""
}

func formRow(form string) entity.FormDataRow {
	key, value, _ := strings.Cut(form, "=")
	if !strings.HasPrefix(value, filePrefix) && !strings.HasPrefix(value, "<") {
		return entity.NewFormDataRow(key, value)
	}
	path, _, _ := strings.Cut(value[1:], ";")
	row := entity.NewFormDataRow(key, "").WithType(formdatatype.File.String())
	row.Src = entity.NewRepickFileRef(path)
	return row
}

func withQuery(address string, parts []string) string {
	if len(parts) == 0 {
		return address
	}
	separator := "?"
	if strings.Contains(address, "?") {
		separator = "&"
	}
	return address + separator + strings.Join(parts, "&")
}

func hasHeader(headers []entity.KeyValue, name string) bool {
	for _, header := range headers {
		if header.HasKey(name) {
			return true
		}
	}
	return false
}

func rawLanguage(contentType, content string) enums.RawLanguage {
	lower := strings.ToLower(contentType)
	switch {
	case strings.Contains(lower, "json"):
		return rawlanguage.Json
	case strings.Contains(lower, "xml"):
		return rawlanguage.Xml
	case strings.Contains(lower, "html"):
		return rawlanguage.Html
	case strings.Contains(lower, "javascript"):
		return rawlanguage.JavaScript
	case lower == "" && json.Valid([]byte(content)) && strings.ContainsAny(strings.TrimSpace(content)[:1], "{["):
		return rawlanguage.Json
	default:
		return rawlanguage.Text
	}
}

func queryEscape(value string) string {
	return strings.ReplaceAll(url.QueryEscape(value), "+", "%20")
}
