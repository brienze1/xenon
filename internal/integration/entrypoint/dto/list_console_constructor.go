package dto

func NewListConsole(filter, search string, limit int) ListConsole {
	return ListConsole{Filter: filter, Search: search, Limit: limit}
}
