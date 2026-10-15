package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

const notOAuthOwner = "the item's auth is not OAuth 2.0"

type oauthOwners struct {
	findTreeItem usecase.FindTreeItem
	findRequest  usecase.FindRequest
	variables    requestVariables
}

func (o oauthOwners) config(ctx context.Context, localId, root, owner string, draft *entity.OAuth2Config) (*entity.OAuth2Config, error) {
	auth, ancestors, found, err := o.owner(ctx, root, owner)
	if err != nil {
		return nil, err
	}
	config := auth.OAuth2
	switch {
	case draft != nil:
		config = *draft
	case !found || !auth.IsOAuth2():
		return nil, usecase.RejectInvalidOAuthConfig(notOAuthOwner)
	}
	scopes, err := o.variables.loadChain(ctx, localId, root, ancestors)
	if err != nil {
		return nil, err
	}
	resolved := config.MapText(func(text string) string {
		return scopes.context.Resolve(text, nil)
	})
	return &resolved, nil
}

func (o oauthOwners) owner(ctx context.Context, root, owner string) (entity.Auth, []entity.Collection, bool, error) {
	owner, aliasId := entity.SplitAuthOwner(owner)
	node, err := o.findTreeItem.FindNode(ctx, root, owner)
	if err != nil {
		return entity.Auth{}, []entity.Collection{}, false, notFoundAsMissing(err)
	}
	ancestors, err := o.findTreeItem.FindAncestors(ctx, root, owner)
	if err != nil {
		return entity.Auth{}, []entity.Collection{}, false, notFoundAsMissing(err)
	}
	if node.IsRequest() {
		request, err := o.findRequest.Find(ctx, root, owner)
		if err != nil {
			return entity.Auth{}, ancestors, false, notFoundAsMissing(err)
		}
		if !request.HasVersion(aliasId) {
			return entity.Auth{}, ancestors, false, nil
		}
		return request.Version(aliasId).Auth, ancestors, true, nil
	}
	container, err := o.findTreeItem.FindContainer(ctx, root, owner)
	if err != nil {
		return entity.Auth{}, ancestors, false, notFoundAsMissing(err)
	}
	return container.Auth, append(ancestors, *container), true, nil
}

func notFoundAsMissing(err error) error {
	if errs.IsNotFoundError(err) {
		return nil
	}
	return err
}
