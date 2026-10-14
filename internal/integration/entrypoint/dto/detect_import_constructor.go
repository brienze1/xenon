package dto

func NewDetectImport(paths []string) DetectImport {
	return DetectImport{Paths: paths}
}
