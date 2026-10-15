package dto

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const checksumFields = 2

type ReleaseChecksums string

func (r ReleaseChecksums) ToEntity() entity.ReleaseChecksums {
	checksums := entity.ReleaseChecksums{}
	for _, line := range strings.Split(string(r), "\n") {
		fields := strings.Fields(line)
		if len(fields) != checksumFields {
			continue
		}
		checksums[strings.TrimPrefix(fields[1], "*")] = strings.ToLower(fields[0])
	}
	return checksums
}
