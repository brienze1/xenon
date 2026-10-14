package dto

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
)

func TestTreeOrdersChildrenAndFallsBackToFileNames(t *testing.T) {
	documents := []entity.ImportDocument{
		{Path: "Nested/.resources/definition.yaml", Data: []byte("$kind: collection\norder: 1000\n")},
		{Path: "Map.request.yaml", Data: []byte("$kind: http-request\norder: 2000\n")},
		{Path: "list.request.yaml", Data: []byte("$kind: http-request\norder: 2000\n")},
		{Path: "Zeta.request.yaml", Data: []byte("$kind: http-request\n")},
		{Path: "Alpha.request.yaml", Data: []byte("$kind: http-request\nname: Alpha/one\n")},
		{Path: ".resources/Map.resources/examples/B.example.yaml", Data: []byte("$kind: http-example\norder: 2\n")},
		{Path: ".resources/Map.resources/examples/A.example.yaml", Data: []byte("$kind: http-example\norder: 1\n")},
	}
	collection, err := NewTree(documents, "", "Root")
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, child := range collection.Children {
		names = append(names, child.Name)
	}
	want := []string{"Nested", "list", "Map", "Alpha/one", "Zeta"}
	if len(names) != len(want) {
		t.Fatalf("got %v, want %v", names, want)
	}
	for index := range want {
		if names[index] != want[index] {
			t.Fatalf("got %v, want %v", names, want)
		}
	}
	examples := collection.Children[2].Request.examples
	if collection.Name != "Root" || len(examples) != 2 || examples[0].Name != "A" || examples[1].Name != "B" {
		t.Fatalf("got collection %q with examples %+v", collection.Name, examples)
	}
}

func TestTreeRefusesAnUnknownKindButSkipsUnsupportedRequests(t *testing.T) {
	skipped, err := NewTree([]entity.ImportDocument{{Path: "S.request.yaml", Data: []byte("$kind: grpc-request\n")}}, "", "Root")
	if err != nil || len(skipped.Children) != 1 || skipped.Children[0].SkippedKind != "grpc-request" {
		t.Fatalf("got %+v, %v", skipped, err)
	}
	_, err = NewTree([]entity.ImportDocument{{Path: "U.request.yaml", Data: []byte("$kind: mystery\n")}}, "", "Root")
	if err == nil || err.Error() != `U.request.yaml is not a valid Postman v3 file: unknown $kind "mystery"` {
		t.Fatalf("got %v", err)
	}
}
