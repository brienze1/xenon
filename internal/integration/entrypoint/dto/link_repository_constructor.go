package dto

func NewLinkRepository(localId, repository, branch, subfolder, accountId, clonePath, mode string) LinkRepository {
	return LinkRepository{
		LocalId:    localId,
		Repository: repository,
		Branch:     branch,
		Subfolder:  subfolder,
		AccountID:  accountId,
		ClonePath:  clonePath,
		Mode:       mode,
	}
}
