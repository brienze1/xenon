package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RestoreWindow struct {
	Screens []Screen `json:"screens" validate:"required,min=1,dive"`
}

func (r *RestoreWindow) ToEntity() []entity.Screen {
	screens := make([]entity.Screen, 0, len(r.Screens))
	for index := range r.Screens {
		screens = append(screens, r.Screens[index].ToEntity())
	}
	return screens
}
