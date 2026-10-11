package dto

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/importorigin"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
)

func compact(t *testing.T, value any) string {
	t.Helper()
	encoded, err := encodeJson(value)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	return string(encoded)
}

func expectJson(t *testing.T, what string, value any, want string) {
	t.Helper()
	if got := compact(t, value); got != want {
		t.Errorf("%s:\n got %s\nwant %s", what, got, want)
	}
}

func xenonRequest() entity.Request {
	request := entity.NewRequest("r1", "Create <order>")
	request.Description = "Creates & returns"
	request.Method = "post"
	request.URL = "https://{{host}}.acme.test:8443/orders/:id?debug=1#top"
	request.Params = []entity.KeyValue{entity.NewKeyValue("debug", "1"), {Key: "off", Value: "x", Description: "unused"}}
	request.PathVariables = []entity.PathVariable{{Key: "id", Value: "7", Description: "order"}}
	request.Headers = []entity.KeyValue{entity.NewKeyValue("Accept", "application/json")}
	request.Body = entity.DefaultBody()
	request.Body.Mode = bodymode.Raw
	request.Body.Raw = entity.BodyRaw{Language: rawlanguage.Json, Content: `{"a":"<b>"}`}
	request.Auth = entity.Auth{Type: authtype.Bearer, Bearer: entity.AuthBearer{Token: "{{token}}"}}
	request.Scripts = entity.Scripts{PreRequest: "a();\nb();", PostResponse: "c();", Disabled: []enums.ScriptPhase{scriptphase.PostResponse}}
	request.Settings.FollowRedirects = false
	request.DisabledAutoHeaders = []string{"User-Agent"}
	return request
}

func importedRequest() entity.Request {
	request := entity.NewRequest("r2", "Upload")
	request.URL = "{{base}}/upload"
	request.Headers = []entity.KeyValue{entity.NewKeyValue("A", "1"), entity.NewKeyValue("B", "2")}
	request.Body = entity.DefaultBody()
	request.Body.Mode = bodymode.FormData
	file := entity.NewFormDataRow("doc", "").WithType("file")
	file.Src = entity.NewRepickFileRef("/a.txt")
	text := entity.NewFormDataRow("kind", "x")
	text.Enabled = false
	request.Body.FormData = []entity.FormDataRow{file, text}
	request.Auth = entity.NewAttributedAuth(authtype.ApiKey, []entity.AuthAttribute{{Key: "key", Value: "K"}, {Key: "value", Value: "{{v}}"}, {Key: "in", Value: "header"}, {Key: "legacy", Value: true}})
	request.Settings.FollowRedirects = true
	example := entity.NewSavedExample("e1", "Ok")
	example.Request.URL = "{{base}}/upload"
	example.Response.Headers = []entity.KeyValue{entity.NewKeyValue("Content-Type", "text/plain")}
	example.Response = example.Response.WithBody([]byte("done"))
	example.Postman = entity.SourceFields{"id": "ex-1", "_postman_previewlanguage": "text", "cookie": []any{}, "responseTime": nil, "header": []any{map[string]any{"type": "text"}}}
	binary := entity.NewSavedExample("e2", "Bin")
	binary.Response = binary.Response.WithBinaryBody([]byte{0xff})
	request.Examples = []entity.SavedExample{example, binary}
	request.Postman = entity.SourceFields{
		"id":                      "req-1",
		"description":             map[string]any{"type": "text/plain"},
		"event":                   map[string]any{"prerequest": map[string]any{"id": "s-1", "packages": map[string]any{}, "type": "text/javascript"}},
		"protocolProfileBehavior": map[string]any{"followRedirects": true, "disableBodyPruning": true},
		"request": map[string]any{
			"header": map[string]any{"raw": true},
			"url":    map[string]any{"host": []any{"ignored"}, "query": []any{}},
			"body":   map[string]any{"formdata": []any{map[string]any{"src": []any{"/a.txt", "/b.txt"}, "contentType": "text/plain"}, map[string]any{"type": "text"}}},
			"auth":   map[string]any{"apikey": map[string]any{"key": "string", "value": "string"}},
			"proxy":  map[string]any{"host": "p"},
		},
	}
	return request
}

func bodyOf(mode string) entity.Body {
	body := entity.DefaultBody()
	switch mode {
	case "urlencoded":
		body.Mode = bodymode.UrlEncoded
		body.UrlEncoded = []entity.KeyValue{entity.NewKeyValue("a", "1")}
	case "binary":
		body.Mode = bodymode.Binary
		body.Binary = entity.FileRef{Src: "files/x.bin"}
	case "graphql":
		body.Mode = bodymode.GraphQl
		body.GraphQl = entity.BodyGraphQl{Query: "{ me }", Variables: "{}"}
	case "text":
		body.Mode = bodymode.Raw
		body.Raw = entity.BodyRaw{Language: rawlanguage.Text, Content: "hi"}
	}
	return body
}

func TestNewItemOfAXenonRequest(t *testing.T) {
	expectJson(t, "request item", NewItem(entity.NewRequestExportNode(xenonRequest(), "A/Create"), true),
		`{"name":"Create <order>","id":"r1","description":"Creates & returns","event":[{"listen":"prerequest","script":{"type":"text/javascript","exec":["a();","b();"]}},{"listen":"test","script":{"type":"text/javascript","exec":["c();"]},"disabled":true}],"request":{"method":"POST","header":[{"key":"Accept","value":"application/json"}],"body":{"mode":"raw","raw":"{\"a\":\"<b>\"}","options":{"raw":{"language":"json"}}},"url":{"raw":"https://{{host}}.acme.test:8443/orders/:id?debug=1#top","protocol":"https","host":["{{host}}","acme","test"],"port":"8443","path":["orders",":id"],"hash":"top","query":[{"key":"debug","value":"1"},{"key":"off","value":"x","disabled":true,"description":"unused"}],"variable":[{"key":"id","value":"7","description":"order"}]},"auth":{"type":"bearer","bearer":[{"key":"token","value":"{{token}}"}]}},"response":[],"protocolProfileBehavior":{"disabledSystemHeaders":{"user-agent":true},"followRedirects":false}}`)
	withoutIds := NewItem(entity.NewRequestExportNode(entity.NewRequest("r9", "Bare"), "A/Bare"), false)
	expectJson(t, "bare request without Xenon ids", withoutIds, `{"name":"Bare","request":{"method":"GET","header":[],"url":{"raw":""}},"response":[]}`)
}

func TestNewItemOfAnImportedRequestMergesThePassthrough(t *testing.T) {
	expectJson(t, "imported item", NewItem(entity.NewRequestExportNode(importedRequest(), "A/Upload"), false),
		`{"name":"Upload","id":"req-1","description":{"content":"","type":"text/plain"},"event":[{"listen":"prerequest","script":{"id":"s-1","type":"text/javascript","exec":[""],"packages":{}}}],"request":{"method":"GET","header":"A: 1\nB: 2","body":{"mode":"formdata","formdata":[{"key":"doc","type":"file","src":["/a.txt","/b.txt"],"contentType":"text/plain"},{"key":"kind","type":"text","value":"x","disabled":true}]},"url":{"raw":"{{base}}/upload","host":["{{base}}"],"path":["upload"]},"auth":{"type":"apikey","apikey":[{"key":"key","value":"K","type":"string"},{"key":"value","value":"{{v}}","type":"string"},{"key":"legacy","value":true}]},"proxy":{"host":"p"}},"response":[{"id":"ex-1","name":"Ok","originalRequest":{"method":"GET","header":[],"url":{"raw":"{{base}}/upload","host":["{{base}}"],"path":["upload"]}},"status":"OK","code":200,"_postman_previewlanguage":"text","header":[{"key":"Content-Type","value":"text/plain","type":"text"}],"cookie":[],"body":"done","responseTime":null},{"name":"Bin","originalRequest":{"method":"GET","header":[],"url":{"raw":""}},"status":"OK","code":200,"header":[],"body":""}],"protocolProfileBehavior":{"disableBodyPruning":true,"followRedirects":true}}`)
}

func TestNewBodyModes(t *testing.T) {
	cases := map[string]string{
		"urlencoded": `{"mode":"urlencoded","urlencoded":[{"key":"a","value":"1"}]}`,
		"binary":     `{"mode":"file","file":{"src":"files/x.bin"}}`,
		"graphql":    `{"mode":"graphql","graphql":{"query":"{ me }","variables":"{}"}}`,
		"text":       `{"mode":"raw","raw":"hi"}`,
	}
	for mode, want := range cases {
		expectJson(t, mode, NewBody(bodyOf(mode), nil), want)
	}
	if body := NewBody(bodyOf("none"), nil); body != nil {
		t.Errorf("a body without content = %+v", body)
	}
	expectJson(t, "disabled body restored", NewBody(entity.DefaultBody(), map[string]any{"mode": "raw", "raw": "ignored", "disabled": true}), `{"disabled":true,"mode":"raw","raw":"ignored"}`)
	expectJson(t, "disabled body edited in Xenon", NewBody(bodyOf("text"), map[string]any{"mode": "raw", "raw": "ignored", "disabled": true}), `{"mode":"raw","raw":"hi"}`)
	expectJson(t, "file extras", NewBody(bodyOf("binary"), map[string]any{"file": map[string]any{"content": ""}, "options": map[string]any{"raw": map[string]any{"headerFamily": "json"}}}), `{"mode":"file","file":{"src":"files/x.bin","content":""},"options":{"raw":{"headerFamily":"json"}}}`)
	expectJson(t, "raw options", NewBody(bodyOf("text"), map[string]any{"options": map[string]any{"raw": map[string]any{"headerFamily": "json"}, "other": 1}}), `{"mode":"raw","raw":"hi","options":{"other":1,"raw":{"headerFamily":"json"}}}`)
	empty := entity.DefaultBody()
	empty.Mode = bodymode.FormData
	expectJson(t, "empty form", NewBody(empty, nil), `{"mode":"formdata","formdata":[]}`)
	empty.Mode = bodymode.UrlEncoded
	expectJson(t, "empty urlencoded", NewBody(empty, nil), `{"mode":"urlencoded","urlencoded":[]}`)
}

func TestNewCollectionAndVariableDocuments(t *testing.T) {
	collection := entity.NewCollection("c1", "Shop")
	collection.Description = "All"
	collection.Auth = entity.Auth{Type: authtype.None}
	collection.Scripts = entity.Scripts{PreRequest: "own();"}
	secret := entity.NewVariableRow("token", "s")
	secret.Secret = true
	number := entity.NewVariableRow("size", "50")
	number.Postman = entity.SourceFields{"type": "number"}
	flag := entity.NewVariableRow("on", "true")
	flag.Postman = entity.SourceFields{"type": "boolean"}
	broken := entity.NewVariableRow("n", "abc")
	broken.Postman = entity.SourceFields{"type": "number"}
	wasSecret := entity.NewVariableRow("old", "x")
	wasSecret.Postman = entity.SourceFields{"type": "secret", "id": "v1"}
	off := entity.NewVariableRow("off", "1")
	off.Enabled = false
	off.Description = "unused"
	collection.Variables = []entity.VariableRow{secret, number, flag, broken, wasSecret, off}
	folder := entity.NewFolder("f1", "Orders")
	root := entity.NewContainerExportNode(collection, "Shop", []entity.ExportNode{entity.NewContainerExportNode(folder, "Shop/Orders", []entity.ExportNode{})}).Rooted(collection, []entity.Scripts{{PreRequest: "inherited();", PostResponse: "after();"}})
	expectJson(t, "collection", NewCollection(root), `{"info":{"_postman_id":"c1","name":"Shop","description":"All","schema":"https://schema.getpostman.com/json/collection/v2.1.0/collection.json"},"item":[{"name":"Orders","id":"f1","item":[]}],"auth":{"type":"noauth"},"event":[{"listen":"prerequest","script":{"type":"text/javascript","exec":["inherited();"]}},{"listen":"test","script":{"type":"text/javascript","exec":["after();"]}},{"listen":"prerequest","script":{"type":"text/javascript","exec":["own();"]}}],"variable":[{"key":"token","value":"","type":"string"},{"key":"size","value":50,"type":"number"},{"key":"on","value":true,"type":"boolean"},{"key":"n","value":"abc","type":"number"},{"key":"old","value":"x","type":"string","id":"v1"},{"key":"off","value":"1","disabled":true,"description":"unused"}]}`)

	imported := entity.NewCollection("c2", "Acme")
	imported.Postman = entity.SourceFields{"info": map[string]any{"_postman_id": "pm-1", "schema": "v2.0", "_exporter_id": "9", "description": map[string]any{"type": "text/markdown"}}, "protocolProfileBehavior": map[string]any{"x": true}, "auth": map[string]any{"bearer": map[string]any{"token": "string"}}}
	imported.Auth = entity.Auth{Type: authtype.Bearer, Bearer: entity.AuthBearer{Token: ""}}
	expectJson(t, "imported collection", NewCollection(entity.NewContainerExportNode(imported, "Acme", []entity.ExportNode{})), `{"info":{"_postman_id":"pm-1","name":"Acme","description":{"content":"","type":"text/markdown"},"schema":"https://schema.getpostman.com/json/collection/v2.1.0/collection.json","_exporter_id":"9"},"item":[],"auth":{"type":"bearer","bearer":[{"key":"token","value":"","type":"string"}]},"protocolProfileBehavior":{"x":true}}`)

	rootFolder := entity.NewFolder("f2", "Users")
	rootFolder.Description = "People"
	rootFolder.Postman = entity.SourceFields{"id": "fld-1", "description": map[string]any{"type": "text/plain"}, "variable": []any{map[string]any{"key": "page"}}, "protocolProfileBehavior": map[string]any{"y": 1}}
	expectJson(t, "folder root", NewCollection(entity.NewContainerExportNode(rootFolder, "Acme/Users", []entity.ExportNode{})), `{"info":{"_postman_id":"fld-1","name":"Users","description":{"content":"People","type":"text/plain"},"schema":"https://schema.getpostman.com/json/collection/v2.1.0/collection.json"},"item":[],"protocolProfileBehavior":{"y":1}}`)

	environment := entity.NewEnvironment("e1", "dev")
	plain := entity.NewVariableRow("base", "http://x")
	plain.Description = "api"
	plain.Postman = entity.SourceFields{"type": "text"}
	hidden := entity.NewVariableRow("key", "")
	hidden.Secret = true
	hidden.Postman = entity.SourceFields{"type": "secret"}
	nowPlain := entity.NewVariableRow("was", "v")
	nowPlain.Postman = entity.SourceFields{"type": "secret"}
	disabled := entity.NewVariableRow("off", "0")
	disabled.Enabled = false
	environment.Variables = []entity.VariableRow{plain, hidden, nowPlain, disabled}
	expectJson(t, "environment", NewEnvironment(environment, "2026-10-11T10:00:00.000Z", "Xenon/1.0.0"), `{"id":"e1","name":"dev","values":[{"key":"base","value":"http://x","type":"text","enabled":true,"description":"api"},{"key":"key","value":"","type":"secret","enabled":true},{"key":"was","value":"v","type":"default","enabled":true},{"key":"off","value":"0","type":"default","enabled":false}],"_postman_variable_scope":"environment","_postman_exported_at":"2026-10-11T10:00:00.000Z","_postman_exported_using":"Xenon/1.0.0"}`)
	environment.Postman = entity.SourceFields{"id": "env-1", "_postman_variable_scope": "environment", "_postman_exported_at": "old", "color": "x"}
	expectJson(t, "imported environment", NewEnvironment(environment, "at", "using"), `{"id":"env-1","name":"dev","values":[{"key":"base","value":"http://x","type":"text","enabled":true,"description":"api"},{"key":"key","value":"","type":"secret","enabled":true},{"key":"was","value":"v","type":"default","enabled":true},{"key":"off","value":"0","type":"default","enabled":false}],"_postman_variable_scope":"environment","_postman_exported_at":"at","_postman_exported_using":"using","color":"x"}`)

	workspace := entity.NewWorkspace("w1", "Acme")
	workspace.Variables = []entity.VariableRow{entity.NewVariableRow("region", "eu")}
	expectJson(t, "globals", NewGlobals(workspace, "at", "using"), `{"id":"w1","name":"Acme Globals","values":[{"key":"region","value":"eu","type":"default","enabled":true}],"_postman_variable_scope":"globals","_postman_exported_at":"at","_postman_exported_using":"using"}`)
	workspace = workspace.WithPostmanGlobals(entity.SourceFields{"id": "g-1", "name": "My Globals", "_postman_variable_scope": "globals"})
	expectJson(t, "imported globals", NewGlobals(workspace, "at", "using"), `{"id":"g-1","name":"My Globals","values":[{"key":"region","value":"eu","type":"default","enabled":true}],"_postman_variable_scope":"globals","_postman_exported_at":"at","_postman_exported_using":"using"}`)
}

func TestNewAuth(t *testing.T) {
	oauth := entity.Auth{Type: authtype.OAuth2, OAuth2: entity.DefaultOAuth2Config()}
	oauth.OAuth2.AccessTokenURL = "https://auth/token"
	oauth.OAuth2.UseToken = entity.UseIDToken
	oauth.OAuth2.TokenRequestParams = []entity.OAuth2Param{{Key: "audience", Value: "api", SendAs: entity.OAuth2SendAsBody, Enabled: true}}
	expectJson(t, "oauth2", NewAuth(oauth, nil), `{"type":"oauth2","oauth2":[{"key":"accessTokenUrl","value":"https://auth/token"},{"key":"tokenRequestParams","value":[{"key":"audience","value":"api","enabled":true,"send_as":"request_body"}]}]}`)
	expectJson(t, "oauth2 postman", NewAuth(oauth, map[string]any{"oauth2": map[string]any{"useToken": "string", "addTokenTo": "string"}}), `{"type":"oauth2","oauth2":[{"key":"accessTokenUrl","value":"https://auth/token"},{"key":"addTokenTo","value":"header","type":"string"},{"key":"tokenRequestParams","value":[{"key":"audience","value":"api","enabled":true,"send_as":"request_body"}]},{"key":"useToken","value":"id_token","type":"string"}]}`)
	digest := entity.Auth{Type: authtype.Digest, Digest: entity.DefaultDigestAuth()}
	digest.Digest.Username, digest.Digest.DisableRetryRequest = "ana", true
	expectJson(t, "digest", NewAuth(digest, nil), `{"type":"digest","digest":[{"key":"username","value":"ana"},{"key":"disableRetryRequest","value":true}]}`)
	aws := entity.Auth{Type: authtype.AwsV4, AwsV4: entity.AwsV4Auth{AccessKey: "AK", Region: "eu", AddAuthDataToQuery: true}}
	expectJson(t, "awsv4", NewAuth(aws, nil), `{"type":"awsv4","awsv4":[{"key":"accessKey","value":"AK"},{"key":"region","value":"eu"},{"key":"addAuthDataToQuery","value":true}]}`)
	basic := entity.Auth{Type: authtype.Basic, Basic: entity.AuthBasic{Username: "u"}}
	expectJson(t, "basic", NewAuth(basic, nil), `{"type":"basic","basic":[{"key":"username","value":"u"}]}`)
	apiKey := entity.Auth{Type: authtype.ApiKey, ApiKey: entity.ApiKeyAuth{Key: "k", In: entity.ApiKeyInQuery}}
	expectJson(t, "apikey", NewAuth(apiKey, nil), `{"type":"apikey","apikey":[{"key":"key","value":"k"},{"key":"in","value":"query"}]}`)
	unsupported := entity.NewUnsupportedAuth(entity.AuthUnsupported{Origin: importorigin.Postman, Type: "hawk", Data: []any{map[string]any{"key": "authId", "value": "h", "type": "string"}}})
	expectJson(t, "hawk", NewAuth(unsupported, nil), `{"type":"hawk","hawk":[{"key":"authId","type":"string","value":"h"}]}`)
	bruno := entity.NewUnsupportedAuth(entity.AuthUnsupported{Origin: importorigin.Bruno, Type: "wsse"})
	expectJson(t, "bruno", NewAuth(bruno, nil), `{"type":"noauth"}`)
	expectJson(t, "none", NewAuth(entity.Auth{Type: authtype.None}, nil), `{"type":"noauth"}`)
	if NewAuth(entity.DefaultAuth(), nil) != nil {
		t.Error("an inherited auth is left out")
	}
}

func TestNewHeaderListAndEmptyRows(t *testing.T) {
	expectJson(t, "no headers", NewHeaderList(nil, nil), `[]`)
	expectJson(t, "raw header block", NewHeaderList([]entity.KeyValue{entity.NewKeyValue("A", "1")}, map[string]any{"raw": true}), `"A: 1"`)
	expectJson(t, "rows keep passthrough by position", NewHeaderList([]entity.KeyValue{entity.NewKeyValue("A", "1"), entity.NewKeyValue("B", "2")}, []any{map[string]any{"type": "text"}}), `[{"key":"A","value":"1","type":"text"},{"key":"B","value":"2"}]`)
}

func TestNewUrlWithoutHost(t *testing.T) {
	expectJson(t, "file url", NewUrl("file:///tmp/x.json", nil, nil, nil), `{"raw":"file:///tmp/x.json","protocol":"file","path":["tmp","x.json"]}`)
	expectJson(t, "bare url", NewUrl("", nil, nil, map[string]any{"host": []any{"h"}, "x": 1}), `{"raw":"","x":1}`)
}

func TestNewFormDataRowSources(t *testing.T) {
	file := entity.NewFormDataRow("doc", "").WithType("file")
	file.Src = entity.FileRef{Src: "/new.txt"}
	expectJson(t, "no passthrough sources", NewFormDataRowOf(file, map[string]any{"src": []any{}}), `{"key":"doc","type":"file","src":"/new.txt"}`)
	expectJson(t, "re-picked file", NewFormDataRowOf(file, map[string]any{"src": []any{"/old.txt", "/b.txt"}}), `{"key":"doc","type":"file","src":"/new.txt"}`)
}

func TestXenonIdsOnlyForXenonCreatedTrees(t *testing.T) {
	request := entity.NewRequest("r", "R")
	request.Postman = entity.SourceFields{"x": 1}
	folder := entity.NewFolder("f", "F")
	root := entity.NewContainerExportNode(entity.NewCollection("c", "C"), "C", []entity.ExportNode{entity.NewContainerExportNode(folder, "C/F", []entity.ExportNode{entity.NewRequestExportNode(request, "C/F/R")})})
	expectJson(t, "imported request below Xenon containers", NewCollection(root), `{"info":{"_postman_id":"c","name":"C","schema":"https://schema.getpostman.com/json/collection/v2.1.0/collection.json"},"item":[{"name":"F","item":[{"name":"R","request":{"method":"GET","header":[],"url":{"raw":""}},"response":[],"x":1}]}]}`)
}

func TestObjectMergesEveryMissingKey(t *testing.T) {
	merged := newObject().set("a", 1).set("nested", newObject().set("x", 1)).merged(map[string]any{"a": 2, "b": 3, "c": 4, "nested": map[string]any{"x": 2, "y": 3}})
	expectJson(t, "merged", merged, `{"a":1,"nested":{"x":1,"y":3},"b":3,"c":4}`)
}

func TestEncodeDocumentIndentsWithTabsWithoutEscapingHtml(t *testing.T) {
	data, err := EncodeDocument(newObject().set("a", "<b>&").set("c", []int{1}))
	if err != nil {
		t.Fatal(err)
	}
	if want := "{\n\t\"a\": \"<b>&\",\n\t\"c\": [\n\t\t1\n\t]\n}\n"; string(data) != want {
		t.Errorf("document = %q, want %q", data, want)
	}
	if _, err := EncodeDocument(map[string]any{"bad": func() {}}); err == nil {
		t.Error("an unencodable document should fail")
	}
}
