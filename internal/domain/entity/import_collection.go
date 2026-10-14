package entity

type ImportCollection struct {
	Secrets  map[string]string
	Warnings []ImportWarning
	Root     ImportNode
}

func NewImportCollection(root ImportNode, warnings []ImportWarning, secrets map[string]string) ImportCollection {
	if secrets == nil {
		secrets = map[string]string{}
	}
	if warnings == nil {
		warnings = []ImportWarning{}
	}
	return ImportCollection{Root: root, Warnings: warnings, Secrets: secrets}
}

func (i ImportCollection) LogValue() any {
	return map[string]any{"name": i.Root.Name(), "warnings": len(i.Warnings), "secrets": len(i.Secrets)}
}
