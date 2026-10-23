package dto

type MCPRunPerformance struct {
	VirtualUsers    *int               `json:"virtual_users" description:"Users, 1-100."`
	DurationSeconds *int               `json:"duration_seconds" description:"Seconds, 1-3600."`
	DataFile        *string            `json:"data_file" description:"CSV or JSON data file path."`
	Profile         *MCPRunLoadProfile `json:"profile" description:"Load profile; 0 picks a default."`
}

func (m *MCPRunPerformance) Over(settings RunPerformanceSettings) RunPerformanceSettings {
	if m == nil {
		return settings
	}
	settings.VirtualUsers = valueOr(m.VirtualUsers, settings.VirtualUsers)
	settings.DurationSeconds = valueOr(m.DurationSeconds, settings.DurationSeconds)
	settings.DataFile = valueOr(m.DataFile, settings.DataFile)
	settings.Profile = m.Profile.Over(settings.Profile)
	return settings
}
