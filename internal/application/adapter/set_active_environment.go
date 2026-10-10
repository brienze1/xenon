package adapter

import "context"

type SetActiveEnvironmentService interface {
	SetActive(ctx context.Context, localId, id string) (string, error)
}
