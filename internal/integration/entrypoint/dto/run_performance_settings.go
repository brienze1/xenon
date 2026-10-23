package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RunPerformanceSettings struct {
	DataFile        string         `json:"data_file"`
	Profile         RunLoadProfile `json:"profile"`
	VirtualUsers    int            `json:"virtual_users"`
	DurationSeconds int            `json:"duration_seconds"`
}

func NewRunPerformanceSettings(settings entity.PerformanceSettings) RunPerformanceSettings {
	return RunPerformanceSettings{DataFile: settings.DataFile, Profile: NewRunLoadProfile(settings.Profile), VirtualUsers: settings.VirtualUsers, DurationSeconds: settings.DurationSeconds}
}

func (r *RunPerformanceSettings) ToEntity() entity.PerformanceSettings {
	return entity.PerformanceSettings{DataFile: r.DataFile, Profile: r.Profile.ToEntity(), VirtualUsers: r.VirtualUsers, DurationSeconds: r.DurationSeconds}
}
