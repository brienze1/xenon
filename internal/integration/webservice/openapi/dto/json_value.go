package dto

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"go.yaml.in/yaml/v3"
)

const jsonIndent = "  "

// exampleNode shows a JSON object or array as YAML; any other text stays a string.
func exampleNode(mediaType, text string) *yaml.Node {
	trimmed := strings.TrimSpace(text)
	if strings.Contains(strings.ToLower(mediaType), "json") && (strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[")) && json.Valid([]byte(trimmed)) {
		decoder := json.NewDecoder(strings.NewReader(trimmed))
		decoder.UseNumber()
		if node, err := jsonNode(decoder); err == nil {
			return node
		}
	}
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: stringTag, Value: text}
}

// exampleText is the example as text: a scalar as written, a structure as indented JSON.
func exampleText(node *yaml.Node) (string, error) {
	node = resolved(node)
	if node == nil {
		return "", nil
	}
	if node.Kind == yaml.ScalarNode {
		return node.Value, nil
	}
	var compact bytes.Buffer
	if err := writeJson(&compact, node); err != nil {
		return "", err
	}
	var indented bytes.Buffer
	if err := json.Indent(&indented, compact.Bytes(), "", jsonIndent); err != nil {
		return "", err
	}
	return indented.String(), nil
}

func jsonNode(decoder *json.Decoder) (*yaml.Node, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	switch value := token.(type) {
	case json.Delim:
		return jsonContainer(decoder, value)
	case json.Number:
		tag := "!!int"
		if strings.ContainsAny(value.String(), ".eE") {
			tag = "!!float"
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: value.String()}, nil
	case string:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: stringTag, Value: value}, nil
	case bool:
		if value {
			return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "true"}, nil
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "false"}, nil
	default:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "null"}, nil
	}
}

func jsonContainer(decoder *json.Decoder, delimiter json.Delim) (*yaml.Node, error) {
	node := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	if delimiter == '{' {
		node = &yaml.Node{Kind: yaml.MappingNode, Tag: mapTag}
	}
	for decoder.More() {
		if node.Kind == yaml.MappingNode {
			key, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			name, _ := key.(string)
			node.Content = append(node.Content, keyNode(name))
		}
		child, err := jsonNode(decoder)
		if err != nil {
			return nil, err
		}
		node.Content = append(node.Content, child)
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	if len(node.Content) == 0 {
		node.Style = yaml.FlowStyle
	}
	return node, nil
}

func writeJson(writer io.Writer, node *yaml.Node) error {
	node = resolved(node)
	switch node.Kind {
	case yaml.MappingNode:
		return writeJsonObject(writer, node)
	case yaml.SequenceNode:
		return writeJsonArray(writer, node)
	case yaml.ScalarNode:
		return writeJsonScalar(writer, node)
	default:
		return errors.New("the example cannot be written as JSON")
	}
}

func writeJsonObject(writer io.Writer, node *yaml.Node) error {
	_, _ = io.WriteString(writer, "{")
	for index := 0; index+1 < len(node.Content); index += 2 {
		if index > 0 {
			_, _ = io.WriteString(writer, ",")
		}
		key, _ := json.Marshal(node.Content[index].Value)
		_, _ = writer.Write(append(key, ':'))
		if err := writeJson(writer, node.Content[index+1]); err != nil {
			return err
		}
	}
	_, err := io.WriteString(writer, "}")
	return err
}

func writeJsonArray(writer io.Writer, node *yaml.Node) error {
	_, _ = io.WriteString(writer, "[")
	for index, child := range node.Content {
		if index > 0 {
			_, _ = io.WriteString(writer, ",")
		}
		if err := writeJson(writer, child); err != nil {
			return err
		}
	}
	_, err := io.WriteString(writer, "]")
	return err
}

func writeJsonScalar(writer io.Writer, node *yaml.Node) error {
	switch node.ShortTag() {
	case "!!int", "!!float":
		if json.Valid([]byte(node.Value)) {
			_, err := io.WriteString(writer, node.Value)
			return err
		}
	case "!!bool":
		var value bool
		if err := node.Decode(&value); err == nil {
			encoded, _ := json.Marshal(value)
			_, err = writer.Write(encoded)
			return err
		}
	case "!!null":
		_, err := io.WriteString(writer, "null")
		return err
	}
	encoded, _ := json.Marshal(node.Value)
	_, err := writer.Write(encoded)
	return err
}
