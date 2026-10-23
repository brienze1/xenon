package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPRunPerformance struct {
	VirtualUsers    *int               `json:"virtual_users" description:"Users, 1-100."`
	DurationSeconds *int               `json:"duration_seconds" description:"Seconds, 1-3600."`
	DataFile        *string            `json:"data_file" description:"CSV or JSON data file path."`
	Profile         *MCPRunLoadProfile `json:"profile" description:"Load profile; 0 picks a default."`
}

func (m *MCPRunPerformance) ToEntity() entity.PerformanceOverrides {
	if m == nil {
		return entity.PerformanceOverrides{}
	}
	return entity.PerformanceOverrides{
		VirtualUsers:    m.VirtualUsers,
		DurationSeconds: m.DurationSeconds,
		DataFile:        m.DataFile,
		Profile:         m.Profile.ToEntity(),
	}
}
