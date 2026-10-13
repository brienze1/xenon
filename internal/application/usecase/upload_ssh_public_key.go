package usecase

import "context"

type UploadSshPublicKey interface {
	UploadKey(ctx context.Context, token, title, publicKey string) error
}
