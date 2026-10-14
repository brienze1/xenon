package dto

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
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

func (c *Conversion) warn(warning entity.ImportWarning) {
	c.warnings = append(c.warnings, warning)
}

func (c *Conversion) skippedRequest(name, kind, parentPath string) {
	c.warn(entity.NewUnsupportedRequestWarning(name, kind, parentPath))
}

func (c *Conversion) extraAuths(owner, path string, entries []AuthEntry) {
	authTypes := make([]string, 0, len(entries))
	for _, entry := range entries {
		authTypes = append(authTypes, entry.Type)
	}
	c.warn(entity.NewMultiAuthWarning(owner, path, authTypes))
}

func (c *Conversion) protocolScript(scriptType, owner, path string) {
	protocol, _, _ := strings.Cut(scriptType, ":")
	if name, known := protocolNames[protocol]; known {
		protocol = name
	}
	c.warn(entity.NewProtocolScriptWarning(scriptType, owner, protocol, path))
}

func (c *Conversion) linkedSecret(key, path string) {
	c.warn(entity.NewLinkedSecretWarning(key, path))
}
