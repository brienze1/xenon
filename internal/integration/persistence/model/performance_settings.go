package model

import "github.com/brienze1/xenon/internal/domain/entity"

type PerformanceSettings struct {
	DataFile        string      `json:"data_file"`
	Profile         LoadProfile `json:"profile"`
	VirtualUsers    int         `json:"virtual_users"`
	DurationSeconds int         `json:"duration_seconds"`
}

func NewPerformanceSettings(settings entity.PerformanceSettings) PerformanceSettings {
	return PerformanceSettings{DataFile: settings.DataFile, Profile: NewLoadProfile(settings.Profile), VirtualUsers: settings.VirtualUsers, DurationSeconds: settings.DurationSeconds}
}

func (p *PerformanceSettings) ToEntity() entity.PerformanceSettings {
	return entity.PerformanceSettings{DataFile: p.DataFile, Profile: p.Profile.ToEntity(), VirtualUsers: p.VirtualUsers, DurationSeconds: p.DurationSeconds}
}
