package entity

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums/httpmethod"
	"github.com/brienze1/xenon/internal/domain/enums/scriptownerkind"
)

var methodToken = regexp.MustCompile(`^[A-Z][A-Z0-9_-]{0,31}$`)

type Request struct {
	Postman             SourceFields
	Bruno               SourceFields
	ID                  string
	Name                string
	Description         string
	Method              string
	URL                 string
	Revision            string
	Body                Body
	Scripts             Scripts
	Params              []KeyValue
	PathVariables       []PathVariable
	Headers             []KeyValue
	DisabledAutoHeaders []string
	Examples            []SavedExample
	Aliases             []RequestAlias
	Auth                Auth
	Settings            RequestSettings
	InheritScripts      bool
	ScriptsEdited       bool
}

func NewRequest(id, name string) Request {
	return Request{
		ID:                  id,
		Name:                strings.TrimSpace(name),
		Method:              httpmethod.Get.String(),
		Params:              []KeyValue{},
		PathVariables:       []PathVariable{},
		Headers:             []KeyValue{},
		DisabledAutoHeaders: []string{},
		Auth:                DefaultAuth(),
		Body:                DefaultBody(),
		Settings:            DefaultRequestSettings(),
		InheritScripts:      true,
	}
}

// WithScriptsOf takes the stored request's scripts unless this request carries scripts edited in the app.
func (r Request) WithScriptsOf(stored Request) Request {
	if r.ScriptsEdited {
		return r
	}
	r.Scripts = stored.Scripts
	r.InheritScripts = stored.InheritScripts
	return r
}

// ScriptOwners lists the owners of the scripts a send runs, outermost first: the ancestors, then the request.
func (r Request) ScriptOwners(ancestors []Collection) []ScriptOwner {
	owners := make([]ScriptOwner, 0, len(ancestors)+1)
	for _, ancestor := range ancestors {
		owners = append(owners, ancestor.ScriptOwner())
	}
	return append(owners, NewScriptOwner(scriptownerkind.Request, r.ID, r.Name, r.Scripts, r.InheritScripts))
}

func (r Request) Location(ancestors []Collection) []string {
	location := make([]string, 0, len(ancestors)+1)
	for _, ancestor := range ancestors {
		location = append(location, ancestor.Name)
	}
	return append(location, r.Name)
}

func (r Request) HasValidName() bool {
	return isValidItemName(r.Name)
}

func (r Request) NormalizedMethod() string {
	return strings.ToUpper(strings.TrimSpace(r.Method))
}

func (r Request) MethodOrDefault() string {
	if method := r.NormalizedMethod(); method != "" {
		return method
	}
	return httpmethod.Get.String()
}

func (r Request) HasID() bool {
	return IsSavedRequestID(r.ID)
}

func IsSavedRequestID(id string) bool {
	return strings.TrimSpace(id) != ""
}

func (r Request) HasValidMethod() bool {
	return methodToken.MatchString(r.NormalizedMethod())
}

func (r Request) CopyAs(id, name string) Request {
	r.ID = id
	r.Name = strings.TrimSpace(name)
	r.Revision = ""
	return r
}

func (r Request) WithFileRefsRelativeTo(workspaceRoot string) Request {
	r.Body = r.Body.RelativeTo(workspaceRoot)
	return r
}

func (r Request) MapText(transform func(string) string) Request {
	r.URL = transform(r.URL)
	r.Params = mapKeyValues(r.Params, transform)
	pathVariables := make([]PathVariable, 0, len(r.PathVariables))
	for _, variable := range r.PathVariables {
		pathVariables = append(pathVariables, variable.MapText(transform))
	}
	r.PathVariables = pathVariables
	r.Headers = mapKeyValues(r.Headers, transform)
	r.Auth = r.Auth.MapText(transform)
	r.Body = r.Body.MapText(transform)
	return r
}

func (r Request) TextFields() []TextField {
	fields := []TextField{NewTextField("url", r.URL)}
	fields = append(fields, keyValueFields("params", r.Params)...)
	for index, variable := range r.PathVariables {
		fields = append(fields, NewTextField(fieldName("path_variables", strconv.Itoa(index), "value"), variable.Value))
	}
	fields = append(fields, keyValueFields("headers", r.Headers)...)
	fields = append(fields, r.Auth.TextFields("auth")...)
	return append(fields, r.Body.TextFields("body")...)
}

func (r Request) InheritedFields(ancestors []Collection) []TextField {
	authOwner := r.inheritedAuthOwner(ancestors)
	fields := make([]TextField, 0)
	for index, ancestor := range ancestors {
		fields = append(fields, ancestor.InheritedFields(index == authOwner, ancestor.InheritableHeaderIndexes(r.overridingHeaders(ancestors[index+1:])))...)
	}
	return fields
}

func (r Request) Inheriting(effective EffectiveRequest) Request {
	r.Auth = effective.Auth
	r.Headers = effective.Headers
	return r
}

func (r Request) inheritedAuthOwner(ancestors []Collection) int {
	owner := -1
	if !r.Auth.IsInherited() {
		return owner
	}
	for index, ancestor := range ancestors {
		if !ancestor.Auth.IsInherited() {
			owner = index
		}
	}
	return owner
}

func (r Request) overridingHeaders(descendants []Collection) []KeyValue {
	overriding := slices.Clone(r.Headers)
	for _, descendant := range descendants {
		overriding = append(overriding, descendant.Headers...)
	}
	return overriding
}

func (r Request) LogValue() any {
	return map[string]any{"id": r.ID, "method": r.Method, "url": r.URL}
}

func (r Request) IsAutoHeaderDisabled(name string) bool {
	for _, disabled := range r.DisabledAutoHeaders {
		if strings.EqualFold(disabled, name) {
			return true
		}
	}
	return false
}

func (r Request) PathVariableKeys() []string {
	keys := make([]string, 0, len(r.PathVariables))
	for _, variable := range r.PathVariables {
		keys = append(keys, variable.Key)
	}
	return keys
}

func (r Request) ShouldEncodeUrl() bool {
	return r.Settings.EncodeUrl
}

func (r Request) Example(id string) (SavedExample, bool) {
	index := r.exampleIndex(id)
	if index < 0 {
		return SavedExample{}, false
	}
	return r.Examples[index], true
}

func (r Request) Identified(id, name string) Request {
	r.ID = id
	r.Name = name
	return r
}

func (r Request) WithExample(example SavedExample) Request {
	examples := slices.Clone(r.Examples)
	if index := r.exampleIndex(example.ID); index >= 0 {
		examples[index] = example
	}
	r.Examples = examples
	return r
}

func (r Request) WithExampleAppended(example SavedExample) Request {
	r.Examples = append(slices.Clone(r.Examples), example)
	return r
}

func (r Request) WithExampleAfter(id string, example SavedExample) Request {
	index := r.exampleIndex(id)
	r.Examples = slices.Insert(slices.Clone(r.Examples), index+1, example)
	return r
}

func (r Request) WithoutExample(id string) Request {
	r.Examples = slices.DeleteFunc(slices.Clone(r.Examples), func(example SavedExample) bool { return example.ID == id })
	return r
}

func (r Request) exampleIndex(id string) int {
	return slices.IndexFunc(r.Examples, func(example SavedExample) bool { return example.ID == id })
}

// SendableOf copies the content a send uses (method, URL, params, path variables, headers, auto-headers, body, auth,
// scripts and settings) from source, keeping everything that identifies the request.
func (r Request) SendableOf(source Request) Request {
	r.Method = source.Method
	r.URL = source.URL
	r.Params = slices.Clone(source.Params)
	r.PathVariables = slices.Clone(source.PathVariables)
	r.Headers = slices.Clone(source.Headers)
	r.DisabledAutoHeaders = slices.Clone(source.DisabledAutoHeaders)
	r.Body = source.Body
	r.Auth = source.Auth
	r.Scripts = source.Scripts
	r.InheritScripts = source.InheritScripts
	r.Settings = source.Settings
	return r
}

// Version is the request as one of its versions: itself for the Default (""), or with an alias's sendable content.
func (r Request) Version(aliasId string) Request {
	alias, found := r.Alias(aliasId)
	if !found {
		return r
	}
	version := r.SendableOf(alias.Version)
	version.ScriptsEdited = false
	return version
}

func (r Request) Alias(id string) (RequestAlias, bool) {
	index := r.aliasIndex(id)
	if index < 0 {
		return RequestAlias{}, false
	}
	return r.Aliases[index], true
}

// HasAliasNamed is true when another alias than exceptId has the name, ignoring case and surrounding spaces.
func (r Request) HasAliasNamed(name, exceptId string) bool {
	return slices.ContainsFunc(r.Aliases, func(alias RequestAlias) bool { return alias.ID != exceptId && alias.HasName(name) })
}

func (r Request) WithAliasAppended(alias RequestAlias) Request {
	r.Aliases = append(slices.Clone(r.Aliases), alias)
	return r
}

func (r Request) WithAlias(alias RequestAlias) Request {
	aliases := slices.Clone(r.Aliases)
	if index := r.aliasIndex(alias.ID); index >= 0 {
		aliases[index] = alias
	}
	r.Aliases = aliases
	return r
}

func (r Request) WithoutAlias(id string) Request {
	r.Aliases = slices.DeleteFunc(slices.Clone(r.Aliases), func(alias RequestAlias) bool { return alias.ID == id })
	return r
}

func (r Request) aliasIndex(id string) int {
	return slices.IndexFunc(r.Aliases, func(alias RequestAlias) bool { return alias.ID == id })
}
