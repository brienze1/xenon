package persistence

import (
	"context"
	"sync"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

func (v *vaultPersistence) FindToken(ctx context.Context, localId, owner string) (*entity.OAuth2Token, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "owner": owner})

	vault, err := v.Find(ctx, localId)
	if err != nil {
		return nil, err
	}
	token, found := vault.TokenOf(owner)
	if !found {
		logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "owner": owner, "found": false})
		return nil, nil
	}

	logger.Debug(ctx, "Finished", token)
	return &token, nil
}

func (v *vaultPersistence) SaveToken(ctx context.Context, localId string, token entity.OAuth2Token) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "token": token})

	v.tokenWrites.Lock()
	defer v.tokenWrites.Unlock()
	vault, err := v.Find(ctx, localId)
	if err != nil {
		return err
	}
	if vault.IsLocked() || vault.IsBroken() {
		return errs.ConflictError("the workspace vault is locked or unreadable; unlock it to keep OAuth 2.0 tokens", "OAU-07409")
	}
	if err := v.Save(ctx, vault.WithToken(token)); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "token": token})
	return nil
}

func (v *vaultPersistence) DeleteToken(ctx context.Context, localId, owner string) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "owner": owner})

	v.tokenWrites.Lock()
	defer v.tokenWrites.Unlock()
	vault, err := v.Find(ctx, localId)
	if err != nil {
		return err
	}
	if _, found := vault.TokenOf(owner); !found {
		logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "owner": owner, "found": false})
		return nil
	}
	if err := v.Save(ctx, vault.WithoutToken(owner)); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "owner": owner})
	return nil
}

func (v *vaultPersistence) LockToken(ctx context.Context, localId, owner string) func() {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "owner": owner})

	lock, _ := v.tokenLocks.LoadOrStore(localId+"/"+owner, &sync.Mutex{})
	mutex := lock.(*sync.Mutex)
	mutex.Lock()

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "owner": owner})
	return mutex.Unlock
}
