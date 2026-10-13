package entity

import "unicode/utf8"

const minPassphraseLength = 8

type PassphraseSetup struct {
	Passphrase   string
	Confirmation string
}

func (p PassphraseSetup) IsConfirmed() bool {
	return p.Passphrase == p.Confirmation
}

func (p PassphraseSetup) IsLongEnough() bool {
	return utf8.RuneCountInString(p.Passphrase) >= minPassphraseLength
}

func (p PassphraseSetup) LogValue() any {
	return map[string]any{"confirmed": p.IsConfirmed(), "longEnough": p.IsLongEnough()}
}
