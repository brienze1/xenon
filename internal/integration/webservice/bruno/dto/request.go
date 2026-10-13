package dto

import (
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/httpmethod"
	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
)

const (
	httpRequest    = "http"
	graphqlRequest = "graphql"
)

var settingKeys = []string{"encodeUrl", "timeout", "followRedirects"}

type Request struct {
	Seq      *int
	Name     string
	Type     string
	Method   string
	Url      string
	Docs     string
	File     string
	Tags     []string
	Params   []Pair
	Path     []Pair
	Headers  []Pair
	Settings []Pair
	Body     Body
	Auth     Auth
	Scripts  Scripts
}

func (r *Request) IsSupported() bool {
	return r.Type == "" || r.Type == httpRequest || r.Type == graphqlRequest
}

func (r *Request) ToEntity(conversion *Conversion, parent entity.Collection, parentPath string) (entity.ImportNode, bool) {
	if !r.IsSupported() {
		message := r.Name + " is a " + r.Type + " request, which Xenon does not support; it was skipped"
		conversion.warn(warningcode.UnsupportedRequestType, message, parent.ID, parentPath)
		return entity.ImportNode{}, false
	}
	path := entity.JoinImportPath(parentPath, r.Name)
	request := entity.NewRequest(conversion.newId(), r.Name)
	request.Method = r.method()
	request.URL = r.Url
	request.Params = keyValues(r.Params)
	for _, pair := range r.Path {
		request.PathVariables = append(request.PathVariables, entity.PathVariable{Key: pair.Key, Value: pair.Value})
	}
	request.Headers = keyValues(r.Headers)
	request.Body = r.Body.ToEntity()
	request.Auth = r.Auth.ToEntity()
	request.Description = r.Docs
	request.Settings = r.settings()
	request.Bruno = r.passthrough()
	conversion.checkAuth(r.Auth, request.ID, path)
	conversion.checkFiles(request.Body, request.ID, path)
	conversion.addScripts(r.Scripts, request.ID, path, r.File)
	return entity.NewRequestImportNode(request), true
}

func (r *Request) method() string {
	if method := strings.ToUpper(strings.TrimSpace(r.Method)); method != "" {
		return method
	}
	return httpmethod.Get.String()
}

func (r *Request) settings() entity.RequestSettings {
	settings := entity.DefaultRequestSettings()
	for _, pair := range r.Settings {
		switch pair.Key {
		case "encodeUrl":
			settings.EncodeUrl = pair.Value != "false"
		case "followRedirects":
			settings.FollowRedirects = pair.Value != "false"
		case "timeout":
			settings.TimeoutMs, _ = strconv.Atoi(pair.Value)
		}
	}
	return settings
}

func (r *Request) passthrough() entity.SourceFields {
	fields := entity.SourceFields(r.Body.Passthrough())
	if r.Seq != nil {
		fields["seq"] = *r.Seq
	}
	if r.Type != "" {
		fields["type"] = r.Type
	}
	if len(r.Tags) > 0 {
		fields["tags"] = r.Tags
	}
	extra := map[string]any{}
	for _, pair := range r.Settings {
		if !slices.Contains(settingKeys, pair.Key) {
			extra[pair.Key] = scalar(pair.Value)
		}
	}
	if len(extra) > 0 {
		fields["settings"] = extra
	}
	return fields
}

func scalar(value string) any {
	if number, err := strconv.Atoi(value); err == nil {
		return number
	}
	if flag, err := strconv.ParseBool(value); err == nil {
		return flag
	}
	return value
}

func nameOrFileName(name, file string) string {
	if strings.TrimSpace(name) != "" {
		return name
	}
	base := path.Base(file)
	return strings.TrimSuffix(base, path.Ext(base))
}
