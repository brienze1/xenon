package usecase

import "context"

type CheckoutBranch interface {
	Checkout(ctx context.Context, clonePath, branch, subfolder string) error
}
