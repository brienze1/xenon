package dto

func NewCloneWorkspace(repository, branch, subfolder, accountId, clonePath, name string) CloneWorkspace {
	return CloneWorkspace{
		Repository: repository,
		Branch:     branch,
		Subfolder:  subfolder,
		AccountID:  accountId,
		ClonePath:  clonePath,
		Name:       name,
	}
}
