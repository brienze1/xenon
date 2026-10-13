package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
)

func FindLiteralCredentials(root entity.ExportNode) []entity.ExportWarning {
	warnings := make([]entity.ExportWarning, 0)
	for _, node := range root.Nodes() {
		for _, field := range node.Auth().CredentialFields() {
			if field.IsLiteral() {
				warnings = append(warnings, entity.NewExportWarning(warningcode.LiteralCredential, node.Path, "The auth field "+field.Name+" holds a typed value, not a variable; it is exported as typed"))
			}
		}
	}
	return warnings
}
