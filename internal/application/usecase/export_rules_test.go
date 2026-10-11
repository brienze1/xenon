package usecase

import (
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodyencoding"
	"github.com/brienze1/xenon/internal/domain/enums/envcolor"
	"github.com/brienze1/xenon/internal/domain/enums/importorigin"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
)

func bearerAuth(token string) entity.Auth {
	return entity.Auth{Type: authtype.Bearer, Bearer: entity.AuthBearer{Token: token}}
}

func scripted(container entity.Collection, pre, post string) entity.Collection {
	container.Scripts = entity.Scripts{PreRequest: pre, PostResponse: post}
	return container
}

func exportRequest(id, name string) entity.Request {
	request := entity.NewRequest(id, name)
	request.Scripts = entity.Scripts{PreRequest: "own()"}
	return request
}

func TestEffectiveExportRootOfACollectionOnlyInlinesHeaders(t *testing.T) {
	collection := entity.NewCollection("c", "Billing")
	collection.Headers = []entity.KeyValue{entity.NewKeyValue("X-Tenant", "acme")}
	request := entity.NewRequest("r", "List")
	request.Headers = []entity.KeyValue{entity.NewKeyValue("Accept", "*/*")}
	node := entity.NewContainerExportNode(collection, "Billing", []entity.ExportNode{entity.NewRequestExportNode(request, "Billing/List")})

	root := EffectiveExportRoot(node, nil)

	if root.Container.ID != "c" || len(root.Container.Headers) != 0 || len(root.InheritedScripts) != 0 {
		t.Fatalf("root = %+v", root.Container)
	}
	headers := root.Children[0].Request.Headers
	if len(headers) != 2 || headers[0].Key != "X-Tenant" || headers[1].Key != "Accept" {
		t.Fatalf("headers = %+v", headers)
	}
}

func TestEffectiveExportRootOfAFolder(t *testing.T) {
	collection := scripted(entity.NewCollection("c", "Billing"), "collectionPre()", "collectionPost()")
	collection.Auth = bearerAuth("{{token}}")
	collection.Variables = []entity.VariableRow{entity.NewVariableRow("base_url", "https://x")}
	parent := scripted(entity.NewFolder("p", "Parent"), "parentPre()", "")
	parent.Scripts.Disabled = nil
	folder := scripted(entity.NewFolder("f", "Invoices"), "folderPre()", "")
	cases := []struct {
		name      string
		folder    entity.Collection
		parent    entity.Collection
		wantAuth  entity.Auth
		inherited []entity.Scripts
	}{
		{name: "inherits auth and scripts", folder: folder, parent: parent, wantAuth: bearerAuth("{{token}}"), inherited: []entity.Scripts{{PreRequest: "collectionPre()", PostResponse: "collectionPost()"}, {PreRequest: "parentPre()"}}},
		{name: "own auth wins", folder: func() entity.Collection { f := folder; f.Auth = entity.Auth{Type: authtype.None}; return f }(), parent: parent, wantAuth: entity.Auth{Type: authtype.None}, inherited: []entity.Scripts{{PreRequest: "collectionPre()", PostResponse: "collectionPost()"}, {PreRequest: "parentPre()"}}},
		{name: "folder cuts inheritance", folder: func() entity.Collection { f := folder; f.InheritScripts = false; return f }(), parent: parent, wantAuth: bearerAuth("{{token}}"), inherited: []entity.Scripts{}},
		{name: "parent cuts inheritance", folder: folder, parent: func() entity.Collection { p := parent; p.InheritScripts = false; return p }(), wantAuth: bearerAuth("{{token}}"), inherited: []entity.Scripts{{PreRequest: "parentPre()"}}},
		{name: "disabled ancestor scripts do not run", folder: folder, parent: func() entity.Collection {
			p := parent
			p.Scripts.Disabled = []enums.ScriptPhase{scriptphase.PreRequest}
			return p
		}(), wantAuth: bearerAuth("{{token}}"), inherited: []entity.Scripts{{PreRequest: "collectionPre()", PostResponse: "collectionPost()"}}},
	}
	for _, tc := range cases {
		node := entity.NewContainerExportNode(tc.folder, "Billing/Parent/Invoices", []entity.ExportNode{})
		root := EffectiveExportRoot(node, []entity.Collection{collection, tc.parent})
		if root.Container.ID != "f" || root.Container.Name != "Invoices" {
			t.Errorf("%s: root = %+v", tc.name, root.Container)
		}
		if !reflect.DeepEqual(root.Container.Auth, tc.wantAuth) {
			t.Errorf("%s: auth = %+v, want %+v", tc.name, root.Container.Auth, tc.wantAuth)
		}
		if !reflect.DeepEqual(root.InheritedScripts, tc.inherited) {
			t.Errorf("%s: inherited = %+v, want %+v", tc.name, root.InheritedScripts, tc.inherited)
		}
		if len(root.Container.Variables) != 1 || root.Container.Variables[0].Key != "base_url" {
			t.Errorf("%s: variables = %+v", tc.name, root.Container.Variables)
		}
	}
}

func TestEffectiveExportRootOfARequest(t *testing.T) {
	collection := scripted(entity.NewCollection("c", "Billing"), "collectionPre()", "")
	collection.Auth = bearerAuth("{{token}}")
	collection.Headers = []entity.KeyValue{entity.NewKeyValue("X-Tenant", "acme"), entity.NewKeyValue("Accept", "text/plain")}
	collection.Variables = []entity.VariableRow{entity.NewVariableRow("base_url", "https://x")}
	folder := entity.NewFolder("f", "Invoices")
	folder.Headers = []entity.KeyValue{entity.NewKeyValue("X-Tenant", "billing")}
	request := exportRequest("r", "Create")
	disabled := entity.NewKeyValue("X-Off", "1")
	disabled.Enabled = false
	request.Headers = []entity.KeyValue{entity.NewKeyValue("accept", "application/json"), disabled}

	root := EffectiveExportRoot(entity.NewRequestExportNode(request, "Billing/Invoices/Create"), []entity.Collection{collection, folder})

	if root.IsRequest() || root.Container.ID != "r" || root.Container.Name != "Create" || root.Container.IsFolder() {
		t.Fatalf("root = %+v", root.Container)
	}
	if !reflect.DeepEqual(root.Container.Auth, bearerAuth("{{token}}")) {
		t.Errorf("auth = %+v", root.Container.Auth)
	}
	if !reflect.DeepEqual(root.InheritedScripts, []entity.Scripts{{PreRequest: "collectionPre()"}}) {
		t.Errorf("inherited = %+v", root.InheritedScripts)
	}
	if len(root.Container.Variables) != 1 || len(root.Children) != 1 {
		t.Fatalf("root = %+v", root)
	}
	exported := root.Children[0].Request
	if exported.ID != "r" || exported.Scripts.PreRequest != "own()" {
		t.Errorf("request = %+v", exported)
	}
	want := []string{"X-Tenant=billing", "accept=application/json", "X-Off=1"}
	got := make([]string, 0)
	for _, header := range exported.Headers {
		got = append(got, header.Key+"="+header.Value)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("headers = %v, want %v", got, want)
	}

	request.Auth = entity.Auth{Type: authtype.None}
	request.InheritScripts = false
	root = EffectiveExportRoot(entity.NewRequestExportNode(request, "Billing/Invoices/Create"), []entity.Collection{collection, folder})
	if !root.Container.Auth.IsInherited() || len(root.InheritedScripts) != 0 {
		t.Errorf("own auth and cut scripts: root = %+v, inherited = %+v", root.Container.Auth, root.InheritedScripts)
	}
	if !root.Children[0].Request.Auth.IsSupported() || root.Children[0].Request.Auth.Type != authtype.None {
		t.Errorf("request auth = %+v", root.Children[0].Request.Auth)
	}
}

func TestEffectiveExportRootWithoutAncestorAuthStaysInherited(t *testing.T) {
	folder := entity.NewFolder("f", "Invoices")
	root := EffectiveExportRoot(entity.NewContainerExportNode(folder, "B/Invoices", []entity.ExportNode{}), []entity.Collection{entity.NewCollection("c", "B")})
	if !root.Container.Auth.IsInherited() {
		t.Errorf("auth = %+v", root.Container.Auth)
	}
}

func TestApplyCurrentValuesAsInitial(t *testing.T) {
	secret := entity.NewVariableRow("token", "")
	secret.Secret = true
	rows := []entity.VariableRow{entity.NewVariableRow("base_url", "https://dev"), entity.NewVariableRow("tenant", "acme"), secret}
	got := ApplyCurrentValuesAsInitial(rows, map[string]string{"base_url": "https://local", "tenant": "", "token": "s3cret"})
	want := []string{"https://local", "acme", ""}
	for index, row := range got {
		if row.InitialValue != want[index] {
			t.Errorf("row %d = %q, want %q", index, row.InitialValue, want[index])
		}
	}
	if rows[0].InitialValue != "https://dev" {
		t.Errorf("the input rows changed: %+v", rows[0])
	}
	if len(ApplyCurrentValuesAsInitial(nil, nil)) != 0 {
		t.Error("no rows should give no rows")
	}
}

func TestFindLiteralCredentials(t *testing.T) {
	collection := entity.NewCollection("c", "Billing")
	collection.Auth = bearerAuth("abc123")
	cases := []struct {
		auth entity.Auth
		want []string
	}{
		{auth: bearerAuth("{{token}}"), want: []string{"Billing"}},
		{auth: bearerAuth("Bearer {{token}}"), want: []string{"Billing"}},
		{auth: bearerAuth(""), want: []string{"Billing"}},
		{auth: bearerAuth("raw-token"), want: []string{"Billing", "Billing/Get"}},
		{auth: entity.Auth{Type: authtype.Basic, Basic: entity.AuthBasic{Username: "admin", Password: "pw"}}, want: []string{"Billing", "Billing/Get"}},
		{auth: entity.Auth{Type: authtype.ApiKey, ApiKey: entity.ApiKeyAuth{Key: "X-Key", Value: "v"}}, want: []string{"Billing", "Billing/Get"}},
		{auth: entity.Auth{Type: authtype.AwsV4, AwsV4: entity.AwsV4Auth{AccessKey: "AK", SecretKey: "{{sk}}", SessionToken: "st"}}, want: []string{"Billing", "Billing/Get"}},
		{auth: entity.Auth{Type: authtype.Digest, Digest: entity.DigestAuth{Password: "pw"}}, want: []string{"Billing", "Billing/Get"}},
		{auth: entity.Auth{Type: authtype.OAuth2, OAuth2: entity.OAuth2Config{ClientSecret: "cs", Password: "{{pw}}"}}, want: []string{"Billing", "Billing/Get"}},
		{auth: entity.DefaultAuth(), want: []string{"Billing"}},
	}
	for _, tc := range cases {
		request := entity.NewRequest("r", "Get")
		request.Auth = tc.auth
		root := entity.NewContainerExportNode(collection, "Billing", []entity.ExportNode{entity.NewRequestExportNode(request, "Billing/Get")})
		warnings := FindLiteralCredentials(root)
		paths := make([]string, 0)
		for _, warning := range warnings {
			paths = append(paths, warning.ItemPath)
			if warning.Code != warningcode.LiteralCredential || warning.Message == "" {
				t.Errorf("warning = %+v", warning)
			}
		}
		if !reflect.DeepEqual(paths, tc.want) {
			t.Errorf("auth %s: paths = %v, want %v", tc.auth.Type, paths, tc.want)
		}
	}
}

func TestFindPostmanDrops(t *testing.T) {
	folder := entity.NewFolder("f", "Invoices")
	folder.InheritScripts = false
	folder.Headers = []entity.KeyValue{entity.NewKeyValue("X-Tenant", "acme")}
	folder.Bruno = entity.SourceFields{"seq": 1}
	folder.Auth = entity.NewUnsupportedAuth(entity.AuthUnsupported{Origin: importorigin.Bruno, Type: "wsse"})
	request := entity.NewRequest("r", "Create")
	request.InheritScripts = false
	request.Settings.TimeoutMs = 500
	request.URL = "{{vault:base}}/x"
	request.Auth = entity.Auth{Type: authtype.OAuth2, OAuth2: entity.DefaultOAuth2Config()}
	request.Auth.OAuth2.UseToken = entity.UseIDToken
	binary := entity.NewSavedExample("e", "Image")
	binary.Response = binary.Response.WithBinaryBody([]byte{0xff})
	text := entity.NewSavedExample("t", "Text")
	request.Examples = []entity.SavedExample{binary, text}
	root := entity.NewContainerExportNode(entity.NewCollection("c", "Billing"), "Billing", []entity.ExportNode{
		entity.NewContainerExportNode(folder, "Billing/Invoices", []entity.ExportNode{entity.NewRequestExportNode(request, "Billing/Invoices/Create")}),
	})

	got := fields(FindPostmanDrops(root))

	want := []string{
		"Billing/Invoices:inheritScripts", "Billing/Invoices:headers", "Billing/Invoices:auth", "Billing/Invoices:bruno",
		"Billing/Invoices/Create:inheritScripts", "Billing/Invoices/Create:settings.timeoutMs", "Billing/Invoices/Create:examples.Image.body",
		"Billing/Invoices/Create:auth.oauth2.useToken, auth.oauth2.autoRefresh", "Billing/Invoices/Create:url",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("drops = %v\nwant %v", got, want)
	}
	if binary.Response.BodyEncoding != bodyencoding.Base64 {
		t.Fatal("the example should be binary")
	}
}

func TestFindPostmanDropsHonoursTheRootInheritToggle(t *testing.T) {
	folder := entity.NewFolder("f", "Invoices")
	folder.InheritScripts = false
	request := entity.NewRequest("r", "Create")
	request.InheritScripts = false
	if drops := FindPostmanDrops(entity.NewContainerExportNode(folder, "B/Invoices", []entity.ExportNode{})); len(drops) != 0 {
		t.Errorf("folder root drops = %+v", drops)
	}
	if drops := FindPostmanDrops(entity.NewRequestExportNode(request, "B/Create")); len(drops) != 0 {
		t.Errorf("request root drops = %+v", drops)
	}
	collection := entity.NewCollection("c", "B")
	collection.InheritScripts = false
	child := entity.NewContainerExportNode(collection, "B", []entity.ExportNode{})
	if drops := FindPostmanDrops(entity.NewContainerExportNode(entity.NewCollection("w", "W"), "W", []entity.ExportNode{child})); len(drops) != 0 {
		t.Errorf("a collection never cuts inheritance: %+v", drops)
	}
}

func TestFindPostmanEnvironmentDrops(t *testing.T) {
	cases := []struct {
		environment entity.Environment
		want        []string
	}{
		{environment: entity.NewEnvironment("e", "dev"), want: []string{}},
		{environment: entity.NewEnvironment("e", "prod"), want: []string{"dev:color", "dev:confirmBeforeSend"}},
		{environment: entity.Environment{Name: "x", Bruno: entity.SourceFields{"a": 1}}, want: []string{"dev:bruno"}},
		{environment: entity.Environment{Name: "x", Color: envcolor.Green}, want: []string{"dev:color"}},
	}
	for _, tc := range cases {
		if got := fields(FindPostmanEnvironmentDrops(tc.environment, "dev")); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: drops = %v, want %v", tc.environment.Name, got, tc.want)
		}
	}
}

func fields(drops []entity.ExportDrop) []string {
	result := make([]string, 0, len(drops))
	for _, drop := range drops {
		if drop.Reason == "" {
			result = append(result, "missing reason")
		}
		result = append(result, drop.ItemPath+":"+drop.Field)
	}
	return result
}
