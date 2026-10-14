package dto

func NewOAuthTokenUse(localId, ownerId, useToken string) OAuthTokenUse {
	return OAuthTokenUse{LocalId: localId, OwnerId: ownerId, UseToken: useToken}
}
