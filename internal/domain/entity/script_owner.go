package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/scriptownerkind"
)

type ScriptOwner struct {
	Kind           enums.ScriptOwnerKind
	ID             string
	Name           string
	Scripts        Scripts
	InheritScripts bool
}

func NewScriptOwner(kind enums.ScriptOwnerKind, id, name string, scripts Scripts, inheritScripts bool) ScriptOwner {
	return ScriptOwner{Kind: kind, ID: id, Name: name, Scripts: scripts, InheritScripts: inheritScripts}
}

func (o ScriptOwner) CutsInheritance() bool {
	return !o.InheritScripts && o.Kind != scriptownerkind.Collection
}

func (o ScriptOwner) Runs(phase enums.ScriptPhase) bool {
	return o.Scripts.HasText(phase) && !o.Scripts.IsDisabled(phase)
}

func (o ScriptOwner) Ref(phase enums.ScriptPhase) ScriptRef {
	return ScriptRef{Phase: phase, OwnerKind: o.Kind, ItemID: o.ID, OwnerName: o.Name, Text: o.Scripts.Text(phase)}
}
