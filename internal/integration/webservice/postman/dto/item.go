package dto

import (
	"encoding/json"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
)

var itemKeys = []string{"name", "description", "item", "request", "response", "event", "auth"}

type Item struct {
	Raw                     map[string]any `json:"-"`
	Request                 *Request       `json:"request"`
	Auth                    *Auth          `json:"auth"`
	ProtocolProfileBehavior map[string]any `json:"protocolProfileBehavior"`
	Name                    string         `json:"name"`
	Description             Description    `json:"description"`
	Items                   []Item         `json:"item"`
	Response                []Response     `json:"response"`
	Event                   []Event        `json:"event"`
	Variable                []KeyValue     `json:"variable"`
}

func (i *Item) UnmarshalJSON(data []byte) error {
	type plain Item
	if err := json.Unmarshal(data, (*plain)(i)); err != nil {
		return err
	}
	i.Raw = decodeRaw(data)
	return nil
}

func (i *Item) IsFolder() bool {
	_, hasItems := i.Raw["item"]
	return hasItems
}

func (i *Item) ToEntity(conversion *Conversion, parentPath string) entity.ImportNode {
	path := entity.JoinImportPath(parentPath, i.Name)
	if i.IsFolder() {
		return i.folder(conversion, path)
	}
	return i.request(conversion, path)
}

func (i *Item) folder(conversion *Conversion, path string) entity.ImportNode {
	folder := entity.NewFolder(conversion.newId(), i.Name)
	folder.Description = i.Description.Content
	folder.Auth = i.Auth.ToEntity()
	folder.Scripts = scriptsToEntity(i.Event)
	folder.Postman = i.passthrough(withEntry(map[string]any{}, "auth", i.Auth.Passthrough()))
	conversion.checkAuth(i.Auth, folder.ID, path)
	if len(i.Variable) > 0 {
		conversion.warn(warningcode.FolderVariables, "Postman folder variables ("+variableNames(i.Variable)+") are kept for export; Xenon has no folder variables", folder.ID, path)
	}
	children := make([]entity.ImportNode, 0, len(i.Items))
	for index := range i.Items {
		children = append(children, i.Items[index].ToEntity(conversion, path))
	}
	return entity.NewContainerImportNode(folder, children)
}

func (i *Item) request(conversion *Conversion, path string) entity.ImportNode {
	request := entity.NewRequest(conversion.newId(), i.Name)
	request.Description = i.Description.Content
	if i.Request != nil {
		request = i.Request.ToEntity(request)
		conversion.checkRequest(i.Request, request.ID, path)
	}
	request.Scripts = scriptsToEntity(i.Event)
	request.Settings = settingsToEntity(i.ProtocolProfileBehavior)
	for index := range i.Response {
		request.Examples = append(request.Examples, i.Response[index].ToEntity(conversion.newId()))
	}
	request.Postman = i.passthrough(i.requestPassthrough())
	if len(i.Variable) > 0 {
		conversion.warn(warningcode.RequestVariables, "Postman request variables ("+variableNames(i.Variable)+") are kept for export; Xenon has no request variables", request.ID, path)
	}
	return entity.NewRequestImportNode(request)
}

func (i *Item) requestPassthrough() map[string]any {
	if i.Request == nil {
		return map[string]any{}
	}
	return withEntry(map[string]any{}, "request", i.Request.Passthrough())
}

func (i *Item) passthrough(extra map[string]any) map[string]any {
	fields := without(i.Raw, itemKeys...)
	fields = withEntry(fields, "description", i.Description.Passthrough())
	fields = withEntry(fields, "event", eventsPassthrough(i.Event))
	for key, value := range extra {
		fields[key] = value
	}
	return fields
}

func settingsToEntity(behavior map[string]any) entity.RequestSettings {
	settings := entity.DefaultRequestSettings()
	if follow, isBool := behavior["followRedirects"].(bool); isBool {
		settings.FollowRedirects = follow
	}
	if strict, isBool := behavior["strictSSL"].(bool); isBool {
		settings.VerifySsl = strict
	}
	if disabled, isBool := behavior["disableUrlEncoding"].(bool); isBool {
		settings.EncodeUrl = !disabled
	}
	if disabled, isBool := behavior["disableCookies"].(bool); isBool {
		settings.DisableCookieJar = disabled
	}
	return settings
}

func variableNames(variables []KeyValue) string {
	names := make([]string, 0, len(variables))
	for _, variable := range variables {
		names = append(names, text(variable.Key))
	}
	return strings.Join(names, ", ")
}

const (
	followRedirectsKey    = "followRedirects"
	strictSslKey          = "strictSSL"
	disableUrlEncodingKey = "disableUrlEncoding"
	disableCookiesKey     = "disableCookies"
	disabledHeadersKey    = "disabledSystemHeaders"
)

func NewItem(node entity.ExportNode, xenonIds bool) Item {
	if node.IsRequest() {
		return newRequestItem(*node.Request, xenonIds)
	}
	container := node.Container
	passthrough := map[string]any(container.Postman)
	return Item{
		Name:        container.Name,
		Description: newDescription(container.Description, passthrough["description"]),
		Items:       NewItems(node.Children, xenonIds),
		Auth:        NewAuth(container.Auth, passthrough["auth"]),
		Event:       NewEvents(container.Scripts, nil, passthrough["event"]),
		Raw:         withEntryValue(without(passthrough, itemKeys...), "id", itemId(passthrough, container.ID, xenonIds)),
	}
}

func NewItems(nodes []entity.ExportNode, xenonIds bool) []Item {
	items := make([]Item, 0, len(nodes))
	for _, node := range nodes {
		items = append(items, NewItem(node, xenonIds))
	}
	return items
}

func newRequestItem(request entity.Request, xenonIds bool) Item {
	passthrough := map[string]any(request.Postman)
	exported := NewRequest(request, passthrough["request"])
	responses := make([]Response, 0, len(request.Examples))
	for _, example := range request.Examples {
		responses = append(responses, NewResponse(example))
	}
	return Item{
		Name:                    request.Name,
		Description:             newDescription(request.Description, passthrough["description"]),
		Request:                 &exported,
		Response:                responses,
		Event:                   NewEvents(request.Scripts, nil, passthrough["event"]),
		ProtocolProfileBehavior: protocolProfileBehavior(request, asMap(passthrough["protocolProfileBehavior"])),
		Raw:                     withEntryValue(without(passthrough, append(itemKeys, "protocolProfileBehavior")...), "id", itemId(passthrough, request.ID, xenonIds)),
	}
}

func (i Item) MarshalJSON() ([]byte, error) {
	item := newObject().set("name", i.Name)
	if id, found := i.Raw["id"]; found {
		item.set("id", id)
	}
	item.setPresent("description", i.Description.Value())
	if len(i.Event) > 0 {
		item.set("event", i.Event)
	}
	if i.Request == nil {
		item.set("item", nonNil(i.Items))
	} else {
		item.set("request", i.Request).set("response", nonNil(i.Response))
	}
	if i.Auth != nil {
		item.set("auth", i.Auth)
	}
	item.setIf(len(i.ProtocolProfileBehavior) > 0, "protocolProfileBehavior", i.ProtocolProfileBehavior)
	return item.merged(i.Raw).MarshalJSON()
}

func protocolProfileBehavior(request entity.Request, passthrough map[string]any) map[string]any {
	behavior := without(passthrough)
	defaults := entity.DefaultRequestSettings()
	settings := map[string]bool{
		followRedirectsKey:    request.Settings.FollowRedirects,
		strictSslKey:          request.Settings.VerifySsl,
		disableUrlEncodingKey: !request.Settings.EncodeUrl,
		disableCookiesKey:     request.Settings.DisableCookieJar,
	}
	fallbacks := map[string]bool{
		followRedirectsKey:    defaults.FollowRedirects,
		strictSslKey:          defaults.VerifySsl,
		disableUrlEncodingKey: !defaults.EncodeUrl,
		disableCookiesKey:     defaults.DisableCookieJar,
	}
	for key, value := range settings {
		if _, fromPostman := passthrough[key]; fromPostman || value != fallbacks[key] {
			behavior[key] = value
		}
	}
	if len(request.DisabledAutoHeaders) > 0 {
		headers := asMap(passthrough[disabledHeadersKey])
		disabled := without(headers)
		for _, name := range request.DisabledAutoHeaders {
			disabled[strings.ToLower(name)] = true
		}
		behavior[disabledHeadersKey] = disabled
	}
	return behavior
}

func withEntryValue(fields map[string]any, key string, value any) map[string]any {
	if value != nil {
		fields[key] = value
	}
	return fields
}
