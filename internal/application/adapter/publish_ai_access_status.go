package adapter

import "context"

type PublishAIAccessStatusService interface {
	Publish(ctx context.Context) error
}
