package entity

type XenonImport struct {
	Root        *ImportNode
	Environment *Environment
	Workspace   *ImportWorkspace
	Secrets     string
}

func (x XenonImport) LogValue() any {
	logged := map[string]any{"hasSecrets": x.Secrets != ""}
	if x.Root != nil {
		logged["root"] = x.Root.Name()
	}
	if x.Environment != nil {
		logged["environment"] = x.Environment.Name
	}
	if x.Workspace != nil {
		logged["workspace"] = x.Workspace.LogValue()
	}
	return logged
}
