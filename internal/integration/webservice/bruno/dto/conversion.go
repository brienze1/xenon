package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
)

type Conversion struct {
	newId    func() string
	warnings []entity.ImportWarning
	scripts  []entity.BrunoScripts
}

func NewConversion(newId func() string) *Conversion {
	return &Conversion{newId: newId, warnings: []entity.ImportWarning{}, scripts: []entity.BrunoScripts{}}
}

func (c *Conversion) Warnings() []entity.ImportWarning {
	return c.warnings
}

func (c *Conversion) Scripts() []entity.BrunoScripts {
	return c.scripts
}

func (c *Conversion) warn(code enums.WarningCode, message, itemId, path string) {
	c.warnings = append(c.warnings, entity.NewImportWarning(code, message, itemId, path))
}

func (c *Conversion) addScripts(scripts Scripts, itemId, path, file string) {
	if converted := scripts.ToEntity(itemId, path, file); !converted.IsEmpty() {
		c.scripts = append(c.scripts, converted)
	}
}

func (c *Conversion) checkAuth(auth Auth, itemId, path string) {
	if auth.ToEntity().IsUnsupported() {
		c.warn(warningcode.UnsupportedAuth, "Auth type "+auth.Mode+" is not supported yet; it is kept for export and the request is sent without auth", itemId, path)
	}
}

func (c *Conversion) checkFiles(body entity.Body, itemId, path string) {
	refs := []entity.FileRef{body.Binary}
	for _, row := range body.FormData {
		refs = append(refs, row.Src)
	}
	for _, ref := range refs {
		if ref.NeedsRepick() {
			c.warn(warningcode.MissingFile, ref.Src+" must be picked again on this machine before it is sent", itemId, path)
		}
	}
}
