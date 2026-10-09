package entity

import (
	"slices"
	"sort"
)

func arrangedNames(listed, present []string) []string {
	order := make([]string, 0, len(present))
	for _, name := range listed {
		if slices.Contains(present, name) && !slices.Contains(order, name) {
			order = append(order, name)
		}
	}
	rest := make([]string, 0, len(present))
	for _, name := range present {
		if !slices.Contains(order, name) {
			rest = append(rest, name)
		}
	}
	sort.Strings(rest)
	return append(order, rest...)
}
