package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
)

type Conversion struct {
	newId    func() string
	warnings []entity.ImportWarning
}

func NewConversion(newId func() string) *Conversion {
	return &Conversion{newId: newId, warnings: []entity.ImportWarning{}}
}

func (c *Conversion) Warnings() []entity.ImportWarning {
	return c.warnings
}

func (c *Conversion) warn(code enums.WarningCode, message, itemId, path string) {
	c.warnings = append(c.warnings, entity.NewImportWarning(code, message, itemId, path))
}

func (c *Conversion) checkAuth(auth *Auth, itemId, path string) {
	if auth.IsUnsupported() {
		c.warn(warningcode.UnsupportedAuth, "Auth type "+auth.Type+" is not supported yet; it is kept for export and the request is sent without auth", itemId, path)
	}
}

func (c *Conversion) checkRequest(request *Request, itemId, path string) {
	c.checkAuth(request.Auth, itemId, path)
	for _, src := range request.Body.FileRefs() {
		c.warn(warningcode.MissingFile, src+" must be picked again on this machine before it is sent", itemId, path)
	}
	if request.Header.IsRawString {
		c.warn(warningcode.RawHeaderString, "The headers were one block of text and were split into rows", itemId, path)
	}
	if request.HasDisabledBody() {
		c.warn(warningcode.DisabledBody, "The body was disabled in Postman; it is kept for export and the request is sent without a body", itemId, path)
	}
}
