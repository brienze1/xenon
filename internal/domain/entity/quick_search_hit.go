package entity

import "strings"

const (
	DefaultQuickSearchLimit  = 50
	quickSearchPathSeparator = " / "
	unrankedQuickSearchHit   = -1
)

type QuickSearchHit struct {
	ID     string
	Name   string
	Method string
	URL    string
	Path   string
	Rank   int
}

func NewQuickSearchHit(node TreeNode, ancestors []TreeNode, query string) (QuickSearchHit, bool) {
	hit := QuickSearchHit{ID: node.ID, Name: node.Name, Method: node.Method, URL: node.URL, Path: quickSearchPath(ancestors)}
	hit.Rank = hit.rank(strings.ToLower(strings.TrimSpace(query)))
	return hit, node.IsRequest() && hit.Rank != unrankedQuickSearchHit
}

func QuickSearchLimit(limit int) int {
	if limit <= 0 {
		return DefaultQuickSearchLimit
	}
	return limit
}

func (q QuickSearchHit) rank(needle string) int {
	name := strings.ToLower(q.Name)
	switch {
	case needle == "":
		return unrankedQuickSearchHit
	case strings.HasPrefix(name, needle):
		return 0
	case strings.Contains(name, needle):
		return 1
	case strings.Contains(strings.ToLower(q.URL), needle):
		return 2
	case strings.Contains(strings.ToLower(q.Path), needle):
		return 3
	default:
		return unrankedQuickSearchHit
	}
}

func quickSearchPath(ancestors []TreeNode) string {
	names := make([]string, 0, len(ancestors))
	for _, ancestor := range ancestors {
		names = append(names, ancestor.Name)
	}
	return strings.Join(names, quickSearchPathSeparator)
}
