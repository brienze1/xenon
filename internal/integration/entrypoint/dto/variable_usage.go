package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type VariableUsage struct {
	Name         string   `json:"name"`
	Status       string   `json:"status"`
	Scope        string   `json:"scope"`
	DisplayValue string   `json:"display_value"`
	UsedIn       []string `json:"used_in"`
	Secret       bool     `json:"secret"`
}

func NewVariableUsage(usage entity.VariableUsage) VariableUsage {
	return VariableUsage{
		Name:         usage.Resolved.Reference.Name,
		Status:       usage.Resolved.Status.String(),
		Scope:        usage.Resolved.Scope.String(),
		DisplayValue: usage.Resolved.DisplayValue(),
		UsedIn:       append([]string{}, usage.UsedIn...),
		Secret:       usage.Resolved.Secret,
	}
}
