package entity

type ImportCounts struct {
	Collections  int
	Folders      int
	Requests     int
	Examples     int
	Variables    int
	Scripts      int
	Environments int
}

func (i ImportCounts) Add(other ImportCounts) ImportCounts {
	return ImportCounts{
		Collections:  i.Collections + other.Collections,
		Folders:      i.Folders + other.Folders,
		Requests:     i.Requests + other.Requests,
		Examples:     i.Examples + other.Examples,
		Variables:    i.Variables + other.Variables,
		Scripts:      i.Scripts + other.Scripts,
		Environments: i.Environments + other.Environments,
	}
}
