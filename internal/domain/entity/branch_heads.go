package entity

type BranchHeads struct {
	Local  string
	Remote string
	Base   string
	Ahead  int
	Behind int
}

func (b BranchHeads) HasLocalCommits() bool {
	return b.Local != ""
}
