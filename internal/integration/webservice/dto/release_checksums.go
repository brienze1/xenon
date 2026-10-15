package dto

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const checksumFields = 2

func NewReleaseChecksums(content string) entity.ReleaseChecksums {
	checksums := entity.ReleaseChecksums{}
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(line)
		if len(fields) != checksumFields {
			continue
		}
		checksums[strings.TrimPrefix(fields[1], "*")] = strings.ToLower(fields[0])
	}
	return checksums
}
