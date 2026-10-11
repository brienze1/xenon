package entity

type ExportOptions struct {
	SuggestedFileName string
	Formats           []ExportFormatOption
	HasSecrets        bool
	HasCurrentValues  bool
}
