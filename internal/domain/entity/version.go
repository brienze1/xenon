package entity

import (
	"strconv"
	"strings"

	"github.com/brienze1/xenon/pkg/errs"
)

const versionParts = 3

type Version struct {
	Major int
	Minor int
	Patch int
}

func ParseVersion(text string) (Version, error) {
	parts := strings.Split(strings.TrimPrefix(strings.TrimSpace(text), "v"), ".")
	if len(parts) != versionParts {
		return Version{}, invalidVersion(text)
	}
	numbers := make([]int, versionParts)
	for index, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 || strconv.Itoa(number) != part {
			return Version{}, invalidVersion(text)
		}
		numbers[index] = number
	}
	return Version{Major: numbers[0], Minor: numbers[1], Patch: numbers[2]}, nil
}

func (v Version) NewerThan(other Version) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	return v.Patch > other.Patch
}

func (v Version) String() string {
	return strconv.Itoa(v.Major) + "." + strconv.Itoa(v.Minor) + "." + strconv.Itoa(v.Patch)
}

func invalidVersion(text string) error {
	return errs.BadRequestError("the version "+text+" is not a release version", "UPD-01422")
}
