package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
)

const notOAuthOwner = "the item's auth is not OAuth 2.0"

type oauthOwners struct {
	findTreeItem usecase.FindTreeItem
	findRequest  usecase.FindRequest
	variables    requestVariables
}

func (o oauthOwners) config(ctx context.Context, localId, root, owner string, draft *entity.OAuth2Config) (*entity.OAuth2Config, error) {
	auth, ancestors, found := o.owner(ctx, root, owner)
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

func (o oauthOwners) owner(ctx context.Context, root, owner string) (entity.Auth, []entity.Collection, bool) {
	node, err := o.findTreeItem.FindNode(ctx, root, owner)
	if err != nil {
		return entity.Auth{}, []entity.Collection{}, false
	}
	ancestors, err := o.findTreeItem.FindAncestors(ctx, root, owner)
	if err != nil {
		return entity.Auth{}, []entity.Collection{}, false
	}
	if node.IsRequest() {
		request, err := o.findRequest.Find(ctx, root, owner)
		if err != nil {
			return entity.Auth{}, ancestors, false
		}
		return request.Auth, ancestors, true
	}
	container, err := o.findTreeItem.FindContainer(ctx, root, owner)
	if err != nil {
		return entity.Auth{}, ancestors, false
	}
	return container.Auth, append(ancestors, *container), true
}
