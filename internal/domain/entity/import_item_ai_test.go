package entity

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/importitemkind"
)

func TestImportItemIncludedInAIImportSkipsWorkspaces(t *testing.T) {
	if NewWorkspaceImportItem("w", ImportWorkspace{}).IncludedInAIImport() {
		t.Error("a workspace item must not be included")
	}
	if !(ImportItem{Kind: importitemkind.Collection}).IncludedInAIImport() {
		t.Error("a collection item must be included")
	}
}
