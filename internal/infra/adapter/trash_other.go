//go:build !darwin || !cgo

package adapter

import "github.com/Bios-Marcel/wastebasket/v2"

func moveToTrash(path string) error {
	return wastebasket.Trash(path)
}
