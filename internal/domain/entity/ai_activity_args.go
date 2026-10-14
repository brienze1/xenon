package entity

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
)

const (
	AIEnvEditToolName         = "env_edit"
	AIEnvEditAddVarAction     = "add_var"
	AIEnvEditSetCurrentAction = "set_current"
	aiSecretValueArgument     = "secret_value"
	aiPassphraseArgument      = "passphrase"
	aiValueArgument           = "value"
	aiSecretFlagArgument      = "secret"
)

func (a AIActivity) CarriesEnvValue() bool {
	return a.Tool == AIEnvEditToolName && slices.Contains([]string{AIEnvEditAddVarAction, AIEnvEditSetCurrentAction}, a.Action)
}

// RedactedArgs hides secret-carrying argument fields; unparsable arguments are kept as given.
func (a AIActivity) RedactedArgs() string {
	var decoded any
	if err := json.Unmarshal([]byte(a.Args), &decoded); err != nil {
		return a.Args
	}
	if root, isObject := decoded.(map[string]any); isObject && a.CarriesEnvValue() {
		maskAIArgument(root, aiValueArgument)
	}
	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(redactAISecretFields(decoded)); err != nil {
		return a.Args
	}
	return strings.TrimSuffix(encoded.String(), "\n")
}

func redactAISecretFields(node any) any {
	switch value := node.(type) {
	case map[string]any:
		for key, field := range value {
			value[key] = redactAISecretFields(field)
		}
		maskAIArgument(value, aiSecretValueArgument)
		maskAIArgument(value, aiPassphraseArgument)
		if flagged, isBool := value[aiSecretFlagArgument].(bool); isBool && flagged {
			maskAIArgument(value, aiValueArgument)
		}
		return value
	case []any:
		for index, item := range value {
			value[index] = redactAISecretFields(item)
		}
		return value
	default:
		return value
	}
}

func maskAIArgument(arguments map[string]any, key string) {
	if _, found := arguments[key]; found {
		arguments[key] = MaskedValue
	}
}
