package entity

import "strconv"

type ImportReadout struct {
	EncryptedSecrets string
	Items            []ImportItem
	Warnings         []ImportWarning
}

func NewImportReadout(capacity int) ImportReadout {
	return ImportReadout{Items: make([]ImportItem, 0, capacity), Warnings: []ImportWarning{}}
}

func (r *ImportReadout) NextKey() string {
	return strconv.Itoa(len(r.Items))
}

func (r *ImportReadout) Append(item ImportItem) {
	r.Items = append(r.Items, item)
}

func (r *ImportReadout) Warn(warnings []ImportWarning) {
	r.Warnings = append(r.Warnings, warnings...)
}

func (r *ImportReadout) KeepEncryptedSecrets(secrets string) {
	r.EncryptedSecrets = secrets
}
