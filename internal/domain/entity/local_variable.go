package entity

import "encoding/json"

const scriptTypeKey = "$type"

// LocalVariableText is how a script value reads inside {{…}}: strings as they are, special values ($type trees) by their value, anything else JSON-encoded.
func LocalVariableText(value any) string {
	if text, isText := value.(string); isText {
		return text
	}
	if tagged, isTagged := value.(map[string]any); isTagged {
		if kind, found := tagged[scriptTypeKey].(string); found {
			return taggedText(kind, tagged["value"])
		}
	}
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func taggedText(kind string, value any) string {
	if kind == "undefined" {
		return ""
	}
	if text, isText := value.(string); isText {
		return text
	}
	encoded, _ := json.Marshal(value)
	return string(encoded)
}
