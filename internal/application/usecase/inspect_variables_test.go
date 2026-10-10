package usecase

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
	"github.com/brienze1/xenon/internal/domain/enums/referencestatus"
	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
)

func secretRow(key string) entity.VariableRow {
	row := entity.NewVariableRow(key, "")
	row.Secret = true
	return row
}

func inspectionContext(locked bool) entity.VariableContext {
	environment := []entity.VariableRow{
		entity.NewVariableRow("base", "https://api.test"),
		entity.NewVariableRow("auth", "Key {{api_key}}"),
		entity.NewVariableRow("a", "{{b}}"),
		entity.NewVariableRow("b", "{{a}}"),
		entity.NewVariableRow("bare", "{{nobody}}"),
		secretRow("api_key"),
		secretRow("missing"),
	}
	return entity.NewVariableContext(
		entity.NewVariableLayer(variablescope.Environment, environment, nil, map[string]string{"api_key": "s3cr3t"}, locked),
		entity.NewVariableLayer(variablescope.Globals, rows("user", "luis", "base", "from-globals"), nil, nil, false),
	)
}

func describe(inspection entity.VariableInspection) []string {
	described := make([]string, 0, len(inspection.References))
	for _, reference := range inspection.References {
		described = append(described, fmt.Sprintf("%s %d-%d %s %s %s %q", reference.Reference.Field, reference.Reference.Start, reference.Reference.End,
			reference.Reference.Name, reference.Status, reference.Scope, reference.DisplayValue()))
	}
	return described
}

func expectDescribed(t *testing.T, inspection entity.VariableInspection, want ...string) {
	t.Helper()
	got := describe(inspection)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("references:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestInspectReportsStatusScopeAndMaskedValues(t *testing.T) {
	request := entity.NewRequest("", "r")
	request.URL = "{{base}}/{{auth}}/{{api_key}}/{{missing}}/{{a}}/{{bare}}/{{$guid}}/{{$nope}}/{{user}}"
	inspection := InspectRequestVariables(request, nil, inspectionContext(false), []string{"$guid"})
	expectDescribed(t, inspection,
		`url 0-8 base resolved environment "https://api.test"`,
		`url 9-17 auth resolved environment "Key •••••"`,
		`url 18-29 api_key resolved environment "•••••"`,
		`url 30-41 missing missing_secret environment "•••••"`,
		`url 42-47 a unresolved environment ""`,
		`url 48-56 bare resolved environment "{{nobody}}"`,
		`url 57-66 $guid dynamic dynamic ""`,
		`url 67-76 $nope unresolved none ""`,
		`url 77-85 user resolved globals "luis"`,
	)
	locked := InspectRequestVariables(request, nil, inspectionContext(true), nil)
	if locked.References[2].Status != referencestatus.LockedSecret || !locked.References[2].Secret || locked.References[6].Status != referencestatus.Unresolved {
		t.Errorf("locked = %v", describe(locked))
	}
}

func TestInspectListsNestedVariablesOncePerField(t *testing.T) {
	request := entity.NewRequest("", "r")
	request.URL = "{{auth}}{{a}}{{auth}}"
	request.Params = []entity.KeyValue{entity.NewKeyValue("k", "{{api_key}}")}
	inspection := InspectRequestVariables(request, nil, inspectionContext(false), nil)
	got := make([]string, 0, len(inspection.Variables))
	for _, usage := range inspection.Variables {
		got = append(got, usage.Resolved.Reference.Name+"="+string(usage.Resolved.Status)+":"+strings.Join(usage.UsedIn, ","))
	}
	want := "auth=resolved:url|api_key=resolved:url,params.0.value|a=unresolved:url|b=unresolved:url"
	if strings.Join(got, "|") != want {
		t.Errorf("variables = %s\nwant %s", strings.Join(got, "|"), want)
	}
}

func TestInspectCountsJavaScriptOffsets(t *testing.T) {
	request := entity.NewRequest("", "r")
	request.URL = "é😀{{user}}ü{{user}}"
	inspection := InspectRequestVariables(request, nil, inspectionContext(false), nil)
	expectDescribed(t, inspection, `url 3-11 user resolved globals "luis"`, `url 12-20 user resolved globals "luis"`)
}

func TestInspectWalksActiveAuthAndBodyModes(t *testing.T) {
	request := entity.NewRequest("", "r")
	request.PathVariables = []entity.PathVariable{{Key: "id", Value: "{{user}}"}}
	request.Headers = []entity.KeyValue{{Key: "{{user}}", Value: "v"}}
	request.Auth = entity.Auth{Type: authtype.Basic, Basic: entity.AuthBasic{Username: "{{user}}", Password: "{{user}}"}, Bearer: entity.AuthBearer{Token: "{{user}}"}}
	request.Body = entity.Body{Mode: bodymode.FormData, Raw: entity.BodyRaw{Content: "{{user}}"}, FormData: []entity.FormDataRow{
		{Key: "{{user}}", Value: "{{user}}", Type: formdatatype.Text, Enabled: true},
		{Key: "f", Value: "{{ignored}}", Src: entity.FileRef{Src: "{{user}}"}, Type: formdatatype.File, Enabled: true},
	}}
	fields := func(inspection entity.VariableInspection) string {
		names := make([]string, 0)
		for _, reference := range inspection.References {
			names = append(names, reference.Reference.Field)
		}
		return strings.Join(names, ",")
	}
	context := inspectionContext(false)
	if got := fields(InspectRequestVariables(request, nil, context, nil)); got != "path_variables.0.value,headers.0.key,auth.basic.username,auth.basic.password,body.formdata.0.key,body.formdata.0.value,body.formdata.1.src" {
		t.Errorf("fields = %s", got)
	}
	request.Auth.Type = authtype.Bearer
	request.Body = entity.Body{Mode: bodymode.UrlEncoded, UrlEncoded: []entity.KeyValue{entity.NewKeyValue("{{user}}", "{{user}}")}, Binary: entity.FileRef{Src: "{{user}}"}}
	if got := fields(InspectRequestVariables(request, nil, context, nil)); got != "path_variables.0.value,headers.0.key,auth.bearer.token,body.urlencoded.0.key,body.urlencoded.0.value" {
		t.Errorf("fields = %s", got)
	}
	request.Auth.Type = authtype.None
	request.Body = entity.Body{Mode: bodymode.Binary, Binary: entity.FileRef{Src: "{{user}}"}, GraphQl: entity.BodyGraphQl{Query: "{{user}}"}}
	if got := fields(InspectRequestVariables(request, nil, context, nil)); got != "path_variables.0.value,headers.0.key,body.binary" {
		t.Errorf("fields = %s", got)
	}
	request.Body = entity.Body{Mode: bodymode.GraphQl, GraphQl: entity.BodyGraphQl{Query: "{{user}}", Variables: "{{user}}"}}
	if got := fields(InspectRequestVariables(request, nil, context, nil)); got != "path_variables.0.value,headers.0.key,body.graphql.query,body.graphql.variables" {
		t.Errorf("fields = %s", got)
	}
	request.Body = entity.Body{Mode: bodymode.None, Raw: entity.BodyRaw{Content: "{{user}}"}}
	if got := fields(InspectRequestVariables(request, nil, context, nil)); got != "path_variables.0.value,headers.0.key" {
		t.Errorf("fields = %s", got)
	}
}

func TestInspectReportsWhatTheRequestInherits(t *testing.T) {
	root := entity.NewCollection("root", "Root")
	root.Auth = entity.Auth{Type: authtype.Bearer, Bearer: entity.AuthBearer{Token: "{{user}}"}}
	root.Headers = []entity.KeyValue{entity.NewKeyValue("X-Root", "{{user}}"), entity.NewKeyValue("X-Shared", "{{user}}"), {Key: "X-Off", Value: "{{user}}"}, entity.NewKeyValue("X-Mine", "{{user}}")}
	folder := entity.NewFolder("folder", "Folder")
	folder.Auth = entity.Auth{Type: authtype.Bearer, Bearer: entity.AuthBearer{Token: "{{base}}"}}
	folder.Headers = []entity.KeyValue{entity.NewKeyValue("x-shared", "{{base}}"), {Key: "X-Root", Value: "disabled does not override"}}
	request := entity.NewRequest("r", "R")
	request.Headers = []entity.KeyValue{entity.NewKeyValue("x-mine", "own")}
	fields := func(request entity.Request, ancestors ...entity.Collection) string {
		names := make([]string, 0)
		for _, reference := range InspectRequestVariables(request, ancestors, inspectionContext(false), nil).References {
			names = append(names, reference.Reference.Field)
		}
		return strings.Join(names, ",")
	}
	if got := fields(request, root, folder); got != "inherited:root.headers.0.value,inherited:folder.auth.bearer.token,inherited:folder.headers.0.value" {
		t.Errorf("fields = %s", got)
	}
	inheritingFolder := folder
	inheritingFolder.Auth = entity.DefaultAuth()
	if got := fields(request, root, inheritingFolder); got != "inherited:root.auth.bearer.token,inherited:root.headers.0.value,inherited:folder.headers.0.value" {
		t.Errorf("fields = %s", got)
	}
	request.Auth = entity.Auth{Type: authtype.Bearer}
	if got := fields(request, root, folder); got != "inherited:root.headers.0.value,inherited:folder.headers.0.value" {
		t.Errorf("fields with own auth = %s", got)
	}
	request.Auth = entity.Auth{Type: authtype.None}
	if got := fields(request, root); got != "inherited:root.headers.0.value,inherited:root.headers.1.value" {
		t.Errorf("fields with one ancestor = %s", got)
	}
}

func TestSuggestVariables(t *testing.T) {
	describeSuggestions := func(suggestions []entity.VariableSuggestion) string {
		described := make([]string, 0, len(suggestions))
		for _, suggestion := range suggestions {
			described = append(described, fmt.Sprintf("%s/%s/%v/%v/%s", suggestion.Name, suggestion.Scope, suggestion.Secret, suggestion.Dynamic, suggestion.Description))
		}
		return strings.Join(described, " ")
	}
	described := entity.NewVariableRow("base", "x")
	described.Description = "the gateway"
	context := entity.NewVariableContext(
		entity.NewVariableLayer(variablescope.Environment, []entity.VariableRow{described, secretRow("Bearer"), entity.NewVariableRow("$guid", "mine")}, nil, nil, false),
		entity.NewVariableLayer(variablescope.Globals, rows("base", "g", "bare", "b", "other", "o"), nil, nil, false),
	)
	dynamic := []string{"$guid", "$randomInt", "$baseless"}
	cases := map[string]string{
		"ba":  "base/environment/false/false/the gateway bare/globals/false/false/",
		"BE":  "Bearer/environment/true/false/",
		"$":   "$guid/environment/false/false/ $randomInt/dynamic/false/true/ $baseless/dynamic/false/true/",
		"$r":  "$randomInt/dynamic/false/true/",
		"zzz": "",
	}
	for prefix, want := range cases {
		if got := describeSuggestions(SuggestVariables(context, dynamic, prefix)); got != want {
			t.Errorf("SuggestVariables(%q) = %q, want %q", prefix, got, want)
		}
	}
	many := make([]string, 0, 60)
	for index := range 60 {
		many = append(many, fmt.Sprintf("$n%02d", index))
	}
	suggestions := SuggestVariables(entity.NewVariableContext(), many, "$n")
	if len(suggestions) != 50 || suggestions[49].Name != "$n49" {
		t.Errorf("got %d suggestions", len(suggestions))
	}
	if exactly := SuggestVariables(entity.NewVariableContext(), many[:50], ""); len(exactly) != 50 {
		t.Errorf("got %d of 50", len(exactly))
	}
}

func TestResolutionOfTwoHundredReferencesTakesUnderTenMilliseconds(t *testing.T) {
	context := inspectionContext(false)
	request := entity.NewRequest("", "r")
	request.Body = entity.Body{Mode: bodymode.Raw, Raw: entity.BodyRaw{Content: strings.Repeat("{{base}}/{{auth}}/{{user}}/{{missing}}/{{$guid}}", 40)}}
	dynamic := func(string) (string, bool) { return "generated", true }
	ResolveRequestVariables(request, nil, context, dynamic)
	fastest := time.Duration(math.MaxInt64)
	var resolved entity.Request
	for range 5 {
		started := time.Now()
		resolved, _ = ResolveRequestVariables(request, nil, context, dynamic)
		fastest = min(fastest, time.Since(started))
	}
	if fastest >= 10*time.Millisecond {
		t.Errorf("resolution took %s", fastest)
	}
	if strings.Count(resolved.Body.Raw.Content, "generated") != 40 {
		t.Errorf("resolved = %q", resolved.Body.Raw.Content[:80])
	}
}
