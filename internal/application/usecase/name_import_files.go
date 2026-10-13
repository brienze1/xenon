package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

func NameImportFiles(nodes []entity.ImportNode, siblings []string) []entity.ImportNode {
	taken := append([]string{}, siblings...)
	named := make([]entity.ImportNode, 0, len(nodes))
	for _, node := range nodes {
		node = node.WithFileName(DeriveFileName(node.Name(), node.Kind(), taken))
		taken = append(taken, node.FileName)
		if !node.IsRequest() {
			node = node.WithChildren(NameImportFiles(node.Children, []string{}))
		}
		named = append(named, node)
	}
	return named
}
