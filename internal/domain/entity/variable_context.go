package entity

type VariableContext struct {
	Layers []VariableLayer
}

func NewVariableContext(layers ...VariableLayer) VariableContext {
	return VariableContext{Layers: append([]VariableLayer{}, layers...)}
}

func (v VariableContext) Lookup(name string) (ContextVariable, bool) {
	for _, layer := range v.Layers {
		if variable, found := layer.Lookup(name); found {
			return variable, true
		}
	}
	return ContextVariable{}, false
}
