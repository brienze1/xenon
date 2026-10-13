package entity

const (
	DefaultAIPort = 47821
	MinAIPort     = 1024
	MaxAIPort     = 65535
)

type AIAccessSettings struct {
	Port          int
	Enabled       bool
	SkipApprovals bool
}

type AIAccessPatch struct {
	Enabled       *bool
	Port          *int
	SkipApprovals *bool
}

func DefaultAIAccessSettings() AIAccessSettings {
	return AIAccessSettings{Enabled: true}
}

func (a AIAccessSettings) HasChosenPort() bool {
	return a.Port != 0
}

func (a AIAccessSettings) HasValidPort() bool {
	return !a.HasChosenPort() || a.Port >= MinAIPort && a.Port <= MaxAIPort
}

// EffectivePort prefers the saved port; the override only replaces the default.
func (a AIAccessSettings) EffectivePort(override *int) int {
	switch {
	case a.HasChosenPort():
		return a.Port
	case override != nil:
		return *override
	default:
		return DefaultAIPort
	}
}

func (a AIAccessSettings) Merged(patch *AIAccessPatch) AIAccessSettings {
	if patch == nil {
		return a
	}
	if patch.Enabled != nil {
		a.Enabled = *patch.Enabled
	}
	if patch.Port != nil {
		a.Port = *patch.Port
	}
	if patch.SkipApprovals != nil {
		a.SkipApprovals = *patch.SkipApprovals
	}
	return a
}

func (a AIAccessSettings) LogValue() any {
	return map[string]any{"enabled": a.Enabled, "port": a.Port, "skipApprovals": a.SkipApprovals}
}
