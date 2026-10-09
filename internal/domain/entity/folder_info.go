package entity

type FolderInfo struct {
	Path             string
	EntryCount       int
	Exists           bool
	IsDirectory      bool
	HasWorkspaceFile bool
}

func (f FolderInfo) IsEmptyOrAbsent() bool {
	return !f.Exists || (f.IsDirectory && f.EntryCount == 0)
}

func (f FolderInfo) IsWorkspace() bool {
	return f.IsDirectory && f.HasWorkspaceFile
}
