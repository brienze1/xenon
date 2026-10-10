package entity

import "encoding/json"

// LocalVariableText is how a script value reads inside {{…}}: strings as they are, anything else JSON-encoded.
func LocalVariableText(value any) string {
	if text, isText := value.(string); isText {
		return text
	}
	encoded, _ := json.Marshal(value)
	return string(encoded)
}
