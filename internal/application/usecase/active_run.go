package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

type ActiveRun struct {
	Control RunControl
	Run     entity.Run
}
