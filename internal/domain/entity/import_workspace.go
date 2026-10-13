package entity

import "github.com/brienze1/xenon/internal/domain/enums/itemkind"

type ImportWorkspace struct {
	Secrets SecretBundle
	Folder  string
	Name    string
	Counts  ImportCounts
}

func NewImportWorkspace(folder, name string, collections []TreeNode, environments []Environment) ImportWorkspace {
	counts := ImportCounts{}
	for _, collection := range collections {
		counts = counts.Add(treeCounts(collection))
	}
	for _, environment := range WellFormedEnvironments(environments) {
		counts = counts.Add(ImportCounts{Environments: 1, Variables: len(environment.Variables)})
	}
	return ImportWorkspace{Folder: folder, Name: name, Counts: counts, Secrets: NewSecretBundle(nil)}
}

func treeCounts(node TreeNode) ImportCounts {
	if !node.IsItem() {
		return ImportCounts{}
	}
	counts := ImportCounts{}
	switch node.Kind {
	case itemkind.Collection:
		counts.Collections = 1
	case itemkind.Folder:
		counts.Folders = 1
	default:
		counts.Requests = 1
		counts.Examples = len(node.Examples)
	}
	for _, child := range node.Children {
		counts = counts.Add(treeCounts(child))
	}
	return counts
}

func (i ImportWorkspace) WithSecrets(bundle SecretBundle) ImportWorkspace {
	i.Secrets = bundle
	return i
}

func (i ImportWorkspace) LogValue() any {
	return map[string]any{"folder": i.Folder, "name": i.Name, "counts": i.Counts}
}
