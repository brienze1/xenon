package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
)

// EffectiveExportRoot makes an exported folder or request the root of a Postman collection carrying the auth, scripts, variables and headers it inherits.
func EffectiveExportRoot(node entity.ExportNode, ancestors []entity.Collection) entity.ExportNode {
	node = inlineHeaders(node, ancestors)
	if len(ancestors) == 0 {
		return node
	}
	inherited := inheritedScripts(node, ancestors)
	if node.IsRequest() {
		root := entity.NewCollection(node.Request.ID, node.Request.Name).WithVariables(ancestors[0].OwnVariables())
		if node.Request.Auth.IsInherited() {
			root.Auth = ancestorAuth(ancestors)
		}
		return entity.NewContainerExportNode(root, node.Path, []entity.ExportNode{node}).Rooted(root, inherited)
	}
	root := node.Container.WithVariables(ancestors[0].OwnVariables())
	if root.Auth.IsInherited() {
		root.Auth = ancestorAuth(ancestors)
	}
	return node.Rooted(root, inherited)
}

func ancestorAuth(ancestors []entity.Collection) entity.Auth {
	for index := len(ancestors) - 1; index >= 0; index-- {
		if !ancestors[index].Auth.IsInherited() {
			return ancestors[index].Auth
		}
	}
	return entity.DefaultAuth()
}

func inheritedScripts(node entity.ExportNode, ancestors []entity.Collection) []entity.Scripts {
	owners := scriptOwners(node, ancestors)
	start := 0
	for index, owner := range owners {
		if owner.CutsInheritance() {
			start = index
		}
	}
	scripts := make([]entity.Scripts, 0, len(ancestors))
	for index := start; index < len(ancestors); index++ {
		if running := runningScripts(owners[index]); !running.IsEmpty() {
			scripts = append(scripts, running)
		}
	}
	return scripts
}

func scriptOwners(node entity.ExportNode, ancestors []entity.Collection) []entity.ScriptOwner {
	if node.IsRequest() {
		return node.Request.ScriptOwners(ancestors)
	}
	owners := make([]entity.ScriptOwner, 0, len(ancestors)+1)
	for _, ancestor := range ancestors {
		owners = append(owners, ancestor.ScriptOwner())
	}
	return append(owners, node.Container.ScriptOwner())
}

func runningScripts(owner entity.ScriptOwner) entity.Scripts {
	running := entity.Scripts{}
	if owner.Runs(scriptphase.PreRequest) {
		running.PreRequest = owner.Scripts.PreRequest
	}
	if owner.Runs(scriptphase.PostResponse) {
		running.PostResponse = owner.Scripts.PostResponse
	}
	return running
}

func inlineHeaders(node entity.ExportNode, chain []entity.Collection) entity.ExportNode {
	if node.IsRequest() {
		request := *node.Request
		request.Headers = withInheritedHeaders(request.Headers, chainHeaders(chain))
		return entity.NewRequestExportNode(request, node.Path)
	}
	below := append(append([]entity.Collection{}, chain...), node.Container)
	children := make([]entity.ExportNode, 0, len(node.Children))
	for _, child := range node.Children {
		children = append(children, inlineHeaders(child, below))
	}
	node.Container.Headers = []entity.KeyValue{}
	return node.WithChildren(children)
}

func chainHeaders(chain []entity.Collection) []entity.KeyValue {
	headers := make([]entity.KeyValue, 0)
	for _, container := range chain {
		headers = overrideHeaders(headers, container.Headers)
	}
	return headers
}

func withInheritedHeaders(own, inherited []entity.KeyValue) []entity.KeyValue {
	headers := make([]entity.KeyValue, 0, len(inherited)+len(own))
	for _, header := range inherited {
		if !overridesHeader(own, header) {
			headers = append(headers, header)
		}
	}
	return append(headers, own...)
}

func overridesHeader(own []entity.KeyValue, header entity.KeyValue) bool {
	for _, row := range own {
		if row.IsActive() && row.HasKey(header.Key) {
			return true
		}
	}
	return false
}
