package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type DynamicValuesWebservice interface {
	usecase.GenerateDynamicValue
	usecase.ListDynamicValueNames
}
