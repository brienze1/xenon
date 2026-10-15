package dto

import "go.yaml.in/yaml/v3"

type Auth struct {
	Raw     any
	Entries []AuthEntry
}

func (a *Auth) UnmarshalYAML(node *yaml.Node) error {
	if err := node.Decode(&a.Raw); err != nil {
		return err
	}
	switch node.Kind {
	case yaml.MappingNode:
		var entry AuthEntry
		if err := node.Decode(&entry); err != nil {
			return err
		}
		a.Entries = []AuthEntry{entry}
	case yaml.SequenceNode:
		return node.Decode(&a.Entries)
	case yaml.ScalarNode:
		if node.Tag != yamlNullTag {
			return NewWrongTypeError(node.Line)
		}
	default:
		return NewWrongTypeError(node.Line)
	}
	return nil
}

func (a *Auth) toPostmanV21(conversion *Conversion, owner, path string, extensions map[string]any) map[string]any {
	if len(a.Entries) == 0 || a.Entries[0].Type == inheritType {
		return nil
	}
	if len(a.Entries) > 1 {
		conversion.extraAuths(owner, path, a.Entries)
		extensions["auth"] = a.Raw
	}
	return a.Entries[0].toPostmanV21()
}
