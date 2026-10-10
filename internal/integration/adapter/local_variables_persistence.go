package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type LocalVariablesPersistence interface {
	usecase.FindLocalVariables
	usecase.SaveLocalVariables
}
