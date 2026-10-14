package dto

import "go.yaml.in/yaml/v3"

const (
	rawBodyMode        = "raw"
	formDataBodyType   = "formdata"
	urlencodedBodyType = "urlencoded"
	graphqlBodyType    = "graphql"
	fileBodyType       = "file"
	noBodyType         = "none"
	textLanguage       = "text"
)

var rawLanguages = []string{"json", "text", "xml", "html", "javascript"}

type Body struct {
	Extra   map[string]any `yaml:",inline"`
	Content yaml.Node      `yaml:"content"`
	Type    string         `yaml:"type"`
}

func (b *Body) toPostmanV21() (map[string]any, error) {
	if b == nil || b.Type == "" || b.Type == noBodyType {
		return nil, nil
	}
	body := copied(b.Extra)
	switch b.Type {
	case formDataBodyType, urlencodedBodyType:
		var rows KeyValues
		if err := b.Content.Decode(&rows); err != nil {
			return nil, err
		}
		body["mode"], body[b.Type] = b.Type, rows.toPostmanV21()
	case graphqlBodyType, fileBodyType:
		var content map[string]any
		if err := b.Content.Decode(&content); err != nil {
			return nil, err
		}
		body["mode"], body[b.Type] = b.Type, content
	default:
		var content any
		if err := b.Content.Decode(&content); err != nil {
			return nil, err
		}
		body["mode"], body["raw"] = rawBodyMode, text(content)
		body["options"] = map[string]any{"raw": map[string]any{"language": rawLanguage(b.Type)}}
	}
	return body, nil
}

func (b *Body) contentText() string {
	if b == nil {
		return ""
	}
	var content any
	if err := b.Content.Decode(&content); err != nil {
		return ""
	}
	return text(content)
}

func rawLanguage(bodyType string) string {
	for _, language := range rawLanguages {
		if language == bodyType {
			return language
		}
	}
	return textLanguage
}
