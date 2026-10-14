package dto

type MCPDocsPlan struct {
	Updated []string `json:"updated"`
	Created []string `json:"created"`
	Deleted []string `json:"deleted"`
}

func NewMCPDocsPlan(plan DocumentationPlan) MCPDocsPlan {
	return MCPDocsPlan{Updated: documentationNames(plan.Updated), Created: documentationNames(plan.Created), Deleted: documentationNames(plan.Deleted)}
}

func documentationNames(items []DocumentationItem) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Name)
	}
	return names
}
