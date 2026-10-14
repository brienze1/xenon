package dto

func NewVariableEdit(key, initialValue, description string, secret bool, enabled *bool, secretValue *string) VariableEdit {
	return VariableEdit{Key: key, InitialValue: initialValue, Description: description, Secret: secret, Enabled: enabled, SecretValue: secretValue}
}
