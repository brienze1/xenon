package dto

func NewOAuthTokenRequest(localId, ownerId string) OAuthTokenRequest {
	return OAuthTokenRequest{LocalId: localId, OwnerId: ownerId}
}
