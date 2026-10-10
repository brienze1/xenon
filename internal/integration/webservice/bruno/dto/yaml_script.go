package dto

import "go.yaml.in/yaml/v3"

type YamlScript struct {
	Type string `yaml:"type"`
	Code string `yaml:"code"`
	Line int    `yaml:"-"`
}

func (y *YamlScript) UnmarshalYAML(node *yaml.Node) error {
	type plain YamlScript
	if err := node.Decode((*plain)(y)); err != nil {
		return err
	}
	for index := 0; index+1 < len(node.Content); index += 2 {
		if code := node.Content[index+1]; node.Content[index].Value == "code" {
			y.Line = code.Line
			if code.Style&(yaml.LiteralStyle|yaml.FoldedStyle) != 0 {
				y.Line++
			}
		}
	}
	return nil
}
