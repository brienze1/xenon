package model

import (
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
)

func TestAliasesAreWrittenAfterExamplesAndRoundTrip(t *testing.T) {
	request := contractRequest()
	alias := entity.NewRequestAlias("01JA2Z7C2D3E4F5G6H7J8K9M0P", "v2", request, nil)
	alias.Version.URL = "https://{{host}}/v2/invoices"
	request = request.WithAliasAppended(alias)
	stored := NewRequest(request)

	written := write(t, stored.FieldsWithAliases())
	if !strings.Contains(written, "aliases:\n  - id: 01JA2Z7C2D3E4F5G6H7J8K9M0P\n    name: v2\n    method: POST\n    url: https://{{host}}/v2/invoices\n") {
		t.Fatalf("aliases not written as expected:\n%s", written)
	}
	if strings.Index(written, "aliases:") > strings.Index(written, "postman:") {
		t.Fatalf("aliases must come before postman:\n%s", written)
	}

	var model Request
	read(t, written, &model)
	got := model.ToEntity()
	if len(got.Aliases) != 1 || got.Aliases[0].Name != "v2" || got.Aliases[0].Version.URL != "https://{{host}}/v2/invoices" {
		t.Fatalf("round trip = %+v", got.Aliases)
	}
}

func TestARequestWithoutAliasesWritesNoAliasesKey(t *testing.T) {
	stored := NewRequest(contractRequest())
	if got := write(t, stored.FieldsWithAliases()); got != contractRequestYaml {
		t.Fatalf("got:\n%s\nwant:\n%s", got, contractRequestYaml)
	}
}

func TestAnAliasLimitedToEnvironmentsWritesThemAfterItsNameAndRoundTrips(t *testing.T) {
	request := contractRequest()
	alias := entity.NewRequestAlias("01JA2Z7C2D3E4F5G6H7J8K9M0P", "v2", request, []string{"01HML", "01DEV"})
	stored := NewRequest(request.WithAliasAppended(alias))

	written := write(t, stored.FieldsWithAliases())
	if !strings.Contains(written, "  - id: 01JA2Z7C2D3E4F5G6H7J8K9M0P\n    name: v2\n    environments:\n      - 01HML\n      - 01DEV\n    method: POST\n") {
		t.Fatalf("environments not written after the name:\n%s", written)
	}

	var model Request
	read(t, written, &model)
	if got := model.ToEntity().Aliases[0].EnvironmentIDs; len(got) != 2 || got[0] != "01HML" || got[1] != "01DEV" {
		t.Fatalf("round trip = %v", got)
	}
}

func TestAnAliasForAllEnvironmentsWritesNoEnvironmentsKey(t *testing.T) {
	request := contractRequest()
	stored := NewRequest(request.WithAliasAppended(entity.NewRequestAlias("01JA2Z7C2D3E4F5G6H7J8K9M0P", "v2", request, nil)))
	if written := write(t, stored.FieldsWithAliases()); strings.Contains(written, "environments") {
		t.Fatalf("an alias for all environments must not write environments:\n%s", written)
	}
}
