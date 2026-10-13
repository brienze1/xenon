package usecase

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	secretValueArg = "secret_value"
	passphraseArg  = "passphrase"
	valueArg       = "value"
	secretFlagArg  = "secret"
	envEditTool    = "env_edit"
)

var valueCarryingEnvActions = map[string]bool{"add_var": true, "set_current": true}

// RedactSecretArgs hides secret-carrying argument fields by tool and field name; unparsable arguments are kept as given.
func RedactSecretArgs(tool, action, args string) string {
	var decoded any
	if err := json.Unmarshal([]byte(args), &decoded); err != nil {
		return args
	}
	if root, isObject := decoded.(map[string]any); isObject && tool == envEditTool && valueCarryingEnvActions[action] {
		if _, found := root[valueArg]; found {
			root[valueArg] = entity.MaskedValue
		}
	}
	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(redactSecretFields(decoded)); err != nil {
		return args
	}
	return strings.TrimSuffix(encoded.String(), "\n")
}

func redactSecretFields(node any) any {
	switch value := node.(type) {
	case map[string]any:
		for key, field := range value {
			value[key] = redactSecretFields(field)
		}
		for _, key := range []string{secretValueArg, passphraseArg} {
			if _, found := value[key]; found {
				value[key] = entity.MaskedValue
			}
		}
		if flagged, isBool := value[secretFlagArg].(bool); isBool && flagged {
			if _, found := value[valueArg]; found {
				value[valueArg] = entity.MaskedValue
			}
		}
		return value
	case []any:
		for index, item := range value {
			value[index] = redactSecretFields(item)
		}
		return value
	default:
		return value
	}
}
