package entity

type ImportProbe struct {
	Name        string
	Text        string
	Schema      string
	Scope       string
	Entries     []string
	IsFolder    bool
	IsPasted    bool
	IsJson      bool
	IsMalformed bool
	HasItems    bool
	HasV1Shape  bool
}
