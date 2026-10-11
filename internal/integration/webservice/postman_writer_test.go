package webservice

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
)

func decodedDocument(t *testing.T, data []byte, err error) map[string]any {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("%s is not JSON: %v", data, err)
	}
	return document
}

func TestPostmanWriterWritesDocuments(t *testing.T) {
	writer := NewPostmanWebservice()
	data, err := writer.WriteCollection(context.Background(), entity.NewContainerExportNode(entity.NewCollection("c", "Shop"), "Shop", []entity.ExportNode{}))
	collection := decodedDocument(t, data, err)
	if info, _ := collection["info"].(map[string]any); info["name"] != "Shop" || info["_postman_id"] != "c" {
		t.Errorf("collection = %v", collection)
	}
	data, err = writer.WriteEnvironment(context.Background(), entity.NewEnvironment("e", "dev"))
	environment := decodedDocument(t, data, err)
	if environment["name"] != "dev" || environment["_postman_variable_scope"] != "environment" || environment["_postman_exported_using"] == "" {
		t.Errorf("environment = %v", environment)
	}
	data, err = writer.WriteGlobals(context.Background(), entity.NewWorkspace("w", "Acme"))
	globals := decodedDocument(t, data, err)
	if globals["name"] != "Acme Globals" || globals["_postman_variable_scope"] != "globals" {
		t.Errorf("globals = %v", globals)
	}
}
