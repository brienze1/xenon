package entity

type ImportFolderListing struct {
	Folder  string
	Inputs  []ImportInput
	Skipped []ImportSkippedEntry
}

func NewImportFolderListing(folder string, skipped []ImportSkippedEntry) ImportFolderListing {
	return ImportFolderListing{Folder: folder, Inputs: make([]ImportInput, 0), Skipped: skipped}
}

func (i *ImportFolderListing) Add(input ImportInput) {
	i.Inputs = append(i.Inputs, input)
}

func (i *ImportFolderListing) Skip(name string, err error) {
	i.Skipped = append(i.Skipped, NewImportSkippedEntry(name, err))
}

func (i *ImportFolderListing) IsEmpty() bool {
	return len(i.Inputs) == 0
}

func (i *ImportFolderListing) HasSkipped() bool {
	return len(i.Skipped) > 0
}

func (i *ImportFolderListing) SkippedWarning() ImportWarning {
	return NewSkippedEntriesWarning(i.Folder, i.Skipped)
}
