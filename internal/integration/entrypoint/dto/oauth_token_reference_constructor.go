package dto

func NewOAuthTokenReference(localId, ownerId string) OAuthTokenReference {
	return OAuthTokenReference{LocalId: localId, OwnerId: ownerId}
}
