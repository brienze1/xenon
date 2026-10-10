package usecase

import (
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
	"github.com/brienze1/xenon/pkg/errs"
)

func rows(pairs ...string) []entity.VariableRow {
	result := make([]entity.VariableRow, 0, len(pairs)/2)
	for index := 0; index < len(pairs); index += 2 {
		result = append(result, entity.NewVariableRow(pairs[index], pairs[index+1]))
	}
	return result
}

func globalsContext(pairs ...string) entity.VariableContext {
	return BuildVariableContext(nil, nil, rows(pairs...), entity.NewLocalVariables(), entity.Vault{})
}

func TestBuildVariableContextPrecedence(t *testing.T) {
	environment := entity.NewEnvironment("e1", "dev")
	environment.Variables = append(rows("tenant", "env", "only_env", "e"), entity.VariableRow{Key: "region", InitialValue: "env-disabled"})
	collection := entity.NewCollection("c1", "Billing")
	collection.Variables = rows("tenant", "collection", "region", "collection-region", "page", "10")
	folder := entity.NewFolder("f1", "Invoices")
	folder.Variables = rows("page", "folder")
	globals := rows("tenant", "globals", "region", "globals-region", "page", "g", "only_global", "g")
	local := entity.NewLocalVariables()
	local.Current["env:e1"] = map[string]string{"only_env": "env-current"}
	local.Current["collection:c1"] = map[string]string{"page": "20"}
	local.Current["globals"] = map[string]string{"only_global": "global-current"}
	cases := []struct {
		name        string
		environment *entity.Environment
		ancestors   []entity.Collection
		want        map[string]string
		scopes      int
	}{
		{"all scopes", &environment, []entity.Collection{collection, folder}, map[string]string{"tenant": "env", "region": "collection-region", "page": "20", "only_env": "env-current", "only_global": "global-current"}, 3},
		{"no environment", nil, []entity.Collection{collection, folder}, map[string]string{"tenant": "collection", "region": "collection-region", "page": "20", "only_env": "{{only_env}}"}, 2},
		{"no collection", &environment, nil, map[string]string{"tenant": "env", "region": "globals-region", "page": "g"}, 2},
		{"folder at the root has no variables", nil, []entity.Collection{folder}, map[string]string{"page": "g", "tenant": "globals"}, 2},
		{"globals only", nil, nil, map[string]string{"tenant": "globals", "only_global": "global-current"}, 1},
	}
	for _, tc := range cases {
		context := BuildVariableContext(tc.environment, tc.ancestors, globals, local, entity.Vault{})
		if len(context.Layers) != tc.scopes {
			t.Errorf("%s: %d layers, want %d", tc.name, len(context.Layers), tc.scopes)
		}
		for name, value := range tc.want {
			if got := ResolveText("{{"+name+"}}", context, nil); got != value {
				t.Errorf("%s: {{%s}} = %q, want %q", tc.name, name, got, value)
			}
		}
	}
	context := BuildVariableContext(&environment, []entity.Collection{collection}, globals, local, entity.Vault{})
	scopes := []any{context.Layers[0].Scope, context.Layers[1].Scope, context.Layers[2].Scope}
	if want := []any{variablescope.Environment, variablescope.Collection, variablescope.Globals}; !reflect.DeepEqual(scopes, want) {
		t.Errorf("layer scopes = %v", scopes)
	}
}

func TestResolveTextRules(t *testing.T) {
	context := globalsContext(
		"name", "plain",
		" spaced ", "with-spaces",
		"Token", "upper",
		"greeting", "hello {{who}}",
		"who", "{{first}} {{last}}",
		"first", "Luis",
		"last", "Brienze",
		"a", "{{b}}",
		"b", "{{a}}",
		"self", "<{{self}}>",
		"same", "{{same}}",
		"dangling", "x-{{nobody}}",
		"$custom", "user",
		"empty", "",
	)
	cases := map[string]string{
		"":                     "",
		"no references":        "no references",
		"{{name}}":             "plain",
		"{{ name }}":           "{{ name }}",
		"{{ spaced }}":         "with-spaces",
		"{{Token}}/{{token}}":  "upper/{{token}}",
		"{{}}":                 "{{}}",
		"{{{name}}}":           "{plain}",
		"{name}":               "{name}",
		"{{name}}{{Token}}":    "plainupper",
		"{{greeting}}!":        "hello Luis Brienze!",
		"{{a}}":                "{{b}}",
		"{{b}}":                "{{a}}",
		"{{same}}":             "{{same}}",
		"{{dangling}}":         "x-{{nobody}}",
		"{{$custom}}":          "user",
		"[{{empty}}]":          "[]",
		"{{missing}}/{{name}}": "{{missing}}/plain",
	}
	for text, want := range cases {
		if got := ResolveText(text, context, nil); got != want {
			t.Errorf("ResolveText(%q) = %q, want %q", text, got, want)
		}
	}
	nineteen := ""
	for index := 0; index < 19; index++ {
		nineteen += "<"
	}
	if got, want := ResolveText("{{self}}", context, nil), nineteen+"{{self}}"+reverse(nineteen); got != want {
		t.Errorf("self reference = %q, want %q", got, want)
	}
}

func reverse(text string) string {
	result := ""
	for range text {
		result += ">"
	}
	return result
}

func TestResolveTextStopsAfterNineteenPasses(t *testing.T) {
	pairs := make([]string, 0, 42)
	for index := 1; index <= 20; index++ {
		pairs = append(pairs, "v"+strconv.Itoa(index), "{{v"+strconv.Itoa(index+1)+"}}")
	}
	pairs = append(pairs, "v21", "end")
	context := globalsContext(pairs...)
	cases := map[string]string{
		"{{v3}}": "end",
		"{{v2}}": "{{v21}}",
		"{{v1}}": "{{v20}}",
	}
	for text, want := range cases {
		if got := ResolveText(text, context, nil); got != want {
			t.Errorf("ResolveText(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestResolveTextDynamicValuesComeLastAndAreFreshPerOccurrence(t *testing.T) {
	calls := 0
	dynamic := func(name string) (string, bool) {
		if name != "$guid" && name != "$nested" {
			return "", false
		}
		calls++
		if name == "$nested" {
			return "{{name}}", true
		}
		return "guid-" + strconv.Itoa(calls), true
	}
	context := globalsContext("name", "plain", "$timestamp", "user-wins")
	cases := map[string]string{
		"{{$guid}}/{{$guid}}": "guid-1/guid-2",
		"{{$timestamp}}":      "user-wins",
		"{{$nope}}":           "{{$nope}}",
		"{{$nested}}":         "plain",
	}
	for text, want := range cases {
		calls = 0
		if got := ResolveText(text, context, dynamic); got != want {
			t.Errorf("ResolveText(%q) = %q, want %q", text, got, want)
		}
	}
	if got := ResolveText("{{$guid}}", context, nil); got != "{{$guid}}" {
		t.Errorf("without a generator dynamic names are unresolved, got %q", got)
	}
}

func TestResolveTextLeavesMissingSecretsLiteralWithoutFallingThrough(t *testing.T) {
	environment := entity.NewEnvironment("e", "dev")
	environment.Variables = []entity.VariableRow{{Key: "api_key", Secret: true, Enabled: true}}
	context := BuildVariableContext(&environment, nil, rows("api_key", "global-plain"), entity.NewLocalVariables(), entity.Vault{})
	dynamic := func(string) (string, bool) { return "generated", true }
	if got := ResolveText("{{api_key}}", context, dynamic); got != "{{api_key}}" {
		t.Errorf("a missing secret is sent literally, got %q", got)
	}
}

func TestResolveRequestVariablesResolvesCopies(t *testing.T) {
	context := globalsContext("base", "https://api.test", "token", "tok", "team", "core", "folder", "f")
	request := entity.NewRequest("r", "Request")
	request.URL = "{{base}}/users"
	request.Headers = []entity.KeyValue{entity.NewKeyValue("X-Team", "{{team}}")}
	request.Auth = entity.Auth{Type: authtype.Bearer, Bearer: entity.AuthBearer{Token: "{{token}}"}}
	collection := entity.NewCollection("c", "C")
	collection.Headers = []entity.KeyValue{entity.NewKeyValue("X-Collection", "{{team}}")}
	folder := entity.NewFolder("f", "F")
	folder.Auth = entity.Auth{Type: authtype.Bearer, Bearer: entity.AuthBearer{Token: "{{folder}}"}}
	resolved, ancestors := ResolveRequestVariables(request, []entity.Collection{collection, folder}, context, nil)
	if resolved.URL != "https://api.test/users" || resolved.Headers[0].Value != "core" || resolved.Auth.Bearer.Token != "tok" {
		t.Errorf("resolved request = %+v", resolved)
	}
	if len(ancestors) != 2 || ancestors[0].Headers[0].Value != "core" || ancestors[1].Auth.Bearer.Token != "f" || ancestors[1].ID != "f" {
		t.Errorf("resolved ancestors = %+v", ancestors)
	}
	if request.URL != "{{base}}/users" || collection.Headers[0].Value != "{{team}}" || folder.Auth.Bearer.Token != "{{folder}}" {
		t.Error("the originals keep the unresolved text")
	}
	calls := 0
	dynamic := func(string) (string, bool) { calls++; return "d", true }
	request.URL = "{{$guid}}"
	if resolvedDynamic, none := ResolveRequestVariables(request, nil, context, dynamic); resolvedDynamic.URL != "d" || none == nil || len(none) != 0 || calls != 1 {
		t.Errorf("dynamic resolution = %q, %v, %d calls", resolvedDynamic.URL, none, calls)
	}
}

func TestCheckSendPreflightAsksTheConfirmingEnvironment(t *testing.T) {
	prod := entity.NewEnvironment("01P", "prod")
	request := entity.NewRequest("", "r")
	preflight := CheckSendPreflight(request, globalsContext(), &prod)
	if !preflight.NeedsConfirmation() || preflight.ConfirmEnvironment.ID != "01P" {
		t.Fatalf("preflight = %+v", preflight)
	}
	var conflict errs.BaseError
	if err := RejectUnconfirmedSend(preflight); !errors.As(err, &conflict) || conflict.Code() != "SND-05409" {
		t.Fatalf("rejection = %v", err)
	}
	want := []errs.ErrorDetails{
		{Attribute: "environment_id", Messages: []string{"01P"}},
		{Attribute: "name", Messages: []string{"prod"}},
		{Attribute: "color", Messages: []string{"red"}},
	}
	if got := conflict.ToDto("").Error.ErrorDetails; !reflect.DeepEqual(got, want) {
		t.Errorf("details = %+v, want %+v", got, want)
	}
	if CheckSendPreflight(request, globalsContext(), nil).NeedsConfirmation() {
		t.Error("no environment needs no confirmation")
	}
}
