package dto

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aitooleffect"
	"github.com/brienze1/xenon/pkg/jsonschema"
)

const (
	alwaysRequired     = "always"
	actionQualifier    = "@"
	actionAlternatives = "|"
	pathSeparator      = "."
	arrayItems         = "[]"
	usedByPrefix       = " Used by "
	requiredForPrefix  = " Required for "
	sentenceEnd        = "."
	listJoiner         = ", "
	lastJoiner         = " and "
)

var attributeIndex = regexp.MustCompile(`\[\d+\]`)

type MCPTool struct {
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	Argument        string             `json:"argument"`
	Actions         []string           `json:"actions"`
	Arguments       []MCPToolArguments `json:"-"`
	RequiresAction  bool               `json:"requires_action"`
	ReadOnly        bool               `json:"read_only"`
	Destructive     bool               `json:"destructive"`
	Idempotent      bool               `json:"idempotent"`
	OpenWorld       bool               `json:"open_world"`
	WorkspaceScoped bool               `json:"workspace_scoped"`
}

type mcpToolArgument struct {
	property    jsonschema.Property
	usedBy      []string
	requiredFor []string
	always      bool
}

type mcpArgumentAlias struct {
	prefix   string
	argument string
	actions  []string
}

func NewMCPTool(tool entity.AITool, arguments ...MCPToolArguments) MCPTool {
	return MCPTool{
		Name:            tool.Name,
		Description:     tool.Description,
		Argument:        tool.ArgumentName(),
		Actions:         tool.Actions,
		RequiresAction:  tool.RequiresAction(),
		ReadOnly:        tool.IsReadOnly(),
		Destructive:     tool.IsDestructive(),
		Idempotent:      tool.Idempotent,
		OpenWorld:       tool.OpenWorld,
		WorkspaceScoped: tool.WorkspaceScoped,
		Arguments:       arguments,
	}
}

func (m *MCPTool) ToEntity() entity.AITool {
	effect := aitooleffect.Write
	if m.ReadOnly {
		effect = aitooleffect.ReadOnly
	}
	if m.Destructive {
		effect = aitooleffect.Destructive
	}
	return entity.AITool{
		Name:            m.Name,
		Description:     m.Description,
		Selector:        m.Argument,
		Actions:         m.Actions,
		Effect:          effect,
		Idempotent:      m.Idempotent,
		OpenWorld:       m.OpenWorld,
		WorkspaceScoped: m.WorkspaceScoped,
	}
}

// InputSchema declares the action selector and every argument the tool's actions decode. It never uses oneOf, anyOf
// or allOf: per-action requirements are written into the argument descriptions.
func (m MCPTool) InputSchema() map[string]any {
	arguments, _ := m.arguments()
	properties := make(map[string]any, len(arguments)+1)
	required := make([]string, 0)
	if len(m.Actions) > 0 {
		properties[m.Argument] = map[string]any{"type": "string", "enum": m.Actions}
		if m.RequiresAction {
			required = append(required, m.Argument)
		}
	}
	always := make([]string, 0)
	for name, argument := range arguments {
		properties[name] = m.describe(argument).Map()
		if argument.always {
			always = append(always, name)
		}
	}
	slices.Sort(always)
	required = append(required, always...)
	schema := map[string]any{"type": "object"}
	if len(properties) > 0 {
		schema["properties"] = properties
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

// ArgumentProperties returns the merged argument properties, or an error when the registrations disagree on a type,
// name an action the tool lacks, or hold a field the schema cannot express.
func (m MCPTool) ArgumentProperties() (map[string]jsonschema.Property, error) {
	arguments, err := m.arguments()
	properties := make(map[string]jsonschema.Property, len(arguments))
	for name, argument := range arguments {
		properties[name] = m.describe(argument)
	}
	return properties, err
}

// ArgumentOf maps a validation attribute of a request DTO back to the argument the client sent for the action.
func (m MCPTool) ArgumentOf(action, attribute string) string {
	normalised := attributeIndex.ReplaceAllString(attribute, arrayItems)
	best := mcpArgumentAlias{}
	for _, alias := range m.aliases() {
		if !alias.appliesTo(action) || !hasPathPrefix(normalised, alias.prefix) || len(alias.prefix) <= len(best.prefix) {
			continue
		}
		best = alias
	}
	if best.prefix == "" {
		return attribute
	}
	return best.argument + normalised[len(best.prefix):]
}

func (m MCPTool) arguments() (map[string]mcpToolArgument, error) {
	arguments := map[string]mcpToolArgument{}
	var problems []string
	for _, registration := range m.Arguments {
		properties, err := jsonschema.Properties(registration.Shape)
		if err != nil {
			problems = append(problems, err.Error())
			continue
		}
		for name, property := range properties {
			if len(property.SchemaFor) > 0 && !slices.Contains(property.SchemaFor, m.Name) {
				continue
			}
			argument, problem := m.argument(registration, property)
			if problem != "" {
				problems = append(problems, name+": "+problem)
			}
			existing, found := arguments[name]
			if !found {
				arguments[name] = argument
				continue
			}
			if !slices.Equal(existing.property.Types, argument.property.Types) {
				problems = append(problems, fmt.Sprintf("%s: %v conflicts with %v", name, existing.property.Types, argument.property.Types))
				continue
			}
			existing.usedBy = append(existing.usedBy, argument.usedBy...)
			existing.requiredFor = append(existing.requiredFor, argument.requiredFor...)
			existing.always = existing.always || argument.always
			arguments[name] = existing
		}
	}
	if len(problems) > 0 {
		slices.Sort(problems)
		return arguments, fmt.Errorf("tool %s: %s", m.Name, strings.Join(problems, "; "))
	}
	return arguments, nil
}

func (m MCPTool) argument(registration MCPToolArguments, property jsonschema.Property) (mcpToolArgument, string) {
	argument := mcpToolArgument{property: property, usedBy: registration.Actions}
	problem := ""
	if len(property.UsedBy) > 0 {
		argument.usedBy = intersect(property.UsedBy, registration.Actions)
		if len(intersect(property.UsedBy, m.Actions)) == 0 && len(argument.usedBy) == 0 {
			problem = fmt.Sprintf("used_by %v names no action of the tool", property.UsedBy)
		}
	}
	for _, action := range property.RequiredFor {
		if action == alwaysRequired {
			argument.always = true
			continue
		}
		if slices.Contains(registration.Actions, action) {
			argument.requiredFor = append(argument.requiredFor, action)
		}
	}
	return argument, problem
}

func (m MCPTool) describe(argument mcpToolArgument) jsonschema.Property {
	description := argument.property.Description
	if len(m.Actions) > 1 {
		description += sentence(usedByPrefix, m.inActionOrder(argument.usedBy))
	}
	description += sentence(requiredForPrefix, m.inActionOrder(argument.requiredFor))
	return argument.property.WithDescription(strings.TrimSpace(description))
}

func (m MCPTool) inActionOrder(actions []string) []string {
	ordered := make([]string, 0, len(actions))
	for _, action := range m.Actions {
		if slices.Contains(actions, action) {
			ordered = append(ordered, action)
		}
	}
	return ordered
}

func (m MCPTool) aliases() []mcpArgumentAlias {
	aliases := make([]mcpArgumentAlias, 0)
	for _, registration := range m.Arguments {
		properties, err := jsonschema.Properties(registration.Shape)
		if err != nil {
			continue
		}
		for name, property := range properties {
			aliases = append(aliases, propertyAliases(name, property, registration.Actions)...)
		}
	}
	return aliases
}

func propertyAliases(path string, property jsonschema.Property, actions []string) []mcpArgumentAlias {
	aliases := make([]mcpArgumentAlias, 0)
	for _, entry := range property.ValidatedAs {
		prefix, qualifier, qualified := strings.Cut(entry, actionQualifier)
		scope := actions
		if qualified {
			scope = intersect(strings.Split(qualifier, actionAlternatives), actions)
		}
		aliases = append(aliases, mcpArgumentAlias{prefix: prefix, argument: path, actions: scope})
	}
	if property.Items != nil {
		aliases = append(aliases, propertyAliases(path+arrayItems, *property.Items, actions)...)
	}
	for name, nested := range property.Properties {
		aliases = append(aliases, propertyAliases(path+pathSeparator+name, nested, actions)...)
	}
	return aliases
}

func (a mcpArgumentAlias) appliesTo(action string) bool {
	return len(a.actions) == 0 || slices.Contains(a.actions, action)
}

func hasPathPrefix(attribute, prefix string) bool {
	if !strings.HasPrefix(attribute, prefix) {
		return false
	}
	rest := attribute[len(prefix):]
	return rest == "" || strings.HasPrefix(rest, pathSeparator) || strings.HasPrefix(rest, arrayItems)
}

func sentence(prefix string, actions []string) string {
	if len(actions) == 0 {
		return ""
	}
	if len(actions) == 1 {
		return prefix + actions[0] + sentenceEnd
	}
	return prefix + strings.Join(actions[:len(actions)-1], listJoiner) + lastJoiner + actions[len(actions)-1] + sentenceEnd
}

func intersect(values, allowed []string) []string {
	kept := make([]string, 0, len(values))
	for _, value := range values {
		if slices.Contains(allowed, value) {
			kept = append(kept, value)
		}
	}
	return kept
}
