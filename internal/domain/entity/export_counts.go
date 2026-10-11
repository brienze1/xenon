package entity

type ExportCounts struct {
	Collections  int
	Folders      int
	Requests     int
	Examples     int
	Environments int
	Variables    int
}

func (e ExportCounts) Add(other ExportCounts) ExportCounts {
	return ExportCounts{
		Collections:  e.Collections + other.Collections,
		Folders:      e.Folders + other.Folders,
		Requests:     e.Requests + other.Requests,
		Examples:     e.Examples + other.Examples,
		Environments: e.Environments + other.Environments,
		Variables:    e.Variables + other.Variables,
	}
}
