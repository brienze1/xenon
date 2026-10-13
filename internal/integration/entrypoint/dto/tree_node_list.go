package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type TreeNodeList []TreeNode

func NewTreeNodeList(tree entity.Tree) TreeNodeList {
	list := make(TreeNodeList, 0, len(tree.Collections()))
	for _, collection := range tree.Collections() {
		list = append(list, NewTreeNode(collection))
	}
	return list
}

func (t TreeNodeList) LogValue() any {
	return map[string]any{"collections": len(t)}
}
