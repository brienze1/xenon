package persistence

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
)

const gitAccountsKey = "gitAccounts"

type gitAccountPersistence struct {
	mu sync.Mutex
}

func NewGitAccountPersistence() adapter.GitAccountPersistence {
	return &gitAccountPersistence{}
}

func (g *gitAccountPersistence) FindAll(ctx context.Context) ([]entity.GitAccount, error) {
	logger.Debug(ctx, "Started")

	g.mu.Lock()
	stored, err := readGitAccounts()
	g.mu.Unlock()
	if err != nil {
		return nil, err
	}
	accounts := make([]entity.GitAccount, 0, len(stored))
	for _, account := range stored {
		accounts = append(accounts, account.ToEntity())
	}

	logger.Debug(ctx, "Finished", map[string]any{"accounts": len(accounts)})
	return accounts, nil
}

func (g *gitAccountPersistence) Find(ctx context.Context, id string) (*entity.GitAccount, error) {
	logger.Debug(ctx, "Started", map[string]any{"id": id})

	accounts, err := g.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, account := range accounts {
		if account.ID == id {
			logger.Debug(ctx, "Finished", account)
			return &account, nil
		}
	}
	return nil, usecase.RejectMissingGitAccount(id)
}

func (g *gitAccountPersistence) Save(ctx context.Context, account entity.GitAccount) error {
	logger.Debug(ctx, "Started", account)

	g.mu.Lock()
	defer g.mu.Unlock()
	stored, err := readGitAccounts()
	if err != nil {
		return err
	}
	saved := model.NewGitAccount(account)
	replaced := false
	for index := range stored {
		if stored[index].ID == account.ID {
			stored[index], replaced = saved, true
		}
	}
	if !replaced {
		stored = append(stored, saved)
	}
	if err := writeGitAccounts(stored, "ACC-21500", "ACC-22500"); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", account)
	return nil
}

func (g *gitAccountPersistence) Delete(ctx context.Context, id string) error {
	logger.Debug(ctx, "Started", map[string]any{"id": id})

	g.mu.Lock()
	defer g.mu.Unlock()
	stored, err := readGitAccounts()
	if err != nil {
		return err
	}
	kept := make([]model.GitAccount, 0, len(stored))
	for _, account := range stored {
		if account.ID != id {
			kept = append(kept, account)
		}
	}
	if err := writeGitAccounts(kept, "ACC-23500", "ACC-24500"); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", map[string]any{"id": id})
	return nil
}

func readGitAccounts() ([]model.GitAccount, error) {
	stored, err := readAppJson()
	if err != nil {
		return nil, err
	}
	accounts := make([]model.GitAccount, 0)
	if raw, found := stored[gitAccountsKey]; found {
		_ = json.Unmarshal(raw, &accounts)
	}
	return accounts, nil
}

func writeGitAccounts(accounts []model.GitAccount, dirCode, writeCode string) error {
	stored, err := readAppJson()
	if err != nil {
		return err
	}
	stored[gitAccountsKey], _ = json.Marshal(accounts)
	if err := os.MkdirAll(appDataPath(), appDataDirMode); err != nil {
		return errs.FileError(err, dirCode)
	}
	data, _ := json.MarshalIndent(stored, "", "  ")
	if err := fileutils.WriteFileAtomic(appDataPath(registryFile), append(data, '\n'), appDataFileMode); err != nil {
		return errs.FileError(err, writeCode)
	}
	return nil
}
