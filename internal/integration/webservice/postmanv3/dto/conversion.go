package dto

import (
	"strconv"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
)

var protocolNames = map[string]string{"graphql": "GraphQL", "grpc": "gRPC"}

type Conversion struct {
	warnings []entity.ImportWarning
}

func NewConversion() *Conversion {
	return &Conversion{warnings: []entity.ImportWarning{}}
}

func (c *Conversion) Warnings() []entity.ImportWarning {
	return c.warnings
}

func (c *Conversion) warn(code enums.WarningCode, message, path string) {
	c.warnings = append(c.warnings, entity.NewImportWarning(code, message, "", path))
}

func (c *Conversion) skippedRequest(name, kind, parentPath string) {
	c.warn(warningcode.UnsupportedRequestType, name+" is a "+kind+" request, which Xenon does not support; it was skipped", parentPath)
}

func (c *Conversion) extraAuths(owner, path string, entries []AuthEntry) {
	others := make([]string, 0, len(entries)-1)
	for _, entry := range entries[1:] {
		others = append(others, entry.Type)
	}
	message := "Postman multi-auth on " + owner + " lists " + strconv.Itoa(len(entries)) + " auths; Xenon imported the first (" + entries[0].Type + ") and kept " + strings.Join(others, ", ") + " for export only"
	c.warn(warningcode.UnsupportedAuth, message, path)
}

func (c *Conversion) protocolScript(scriptType, owner, path string) {
	protocol, _, _ := strings.Cut(scriptType, ":")
	if name, known := protocolNames[protocol]; known {
		protocol = name
	}
	c.warn(warningcode.ProtocolScript, "The "+scriptType+" script on "+owner+" runs only for "+protocol+" requests in Postman; it is kept for export and Xenon does not run it", path)
}

func (c *Conversion) linkedSecret(key, path string) {
	c.warn(warningcode.LinkedSecret, key+" is linked to a Postman vault secret; enter its value in Xenon", path)
}
