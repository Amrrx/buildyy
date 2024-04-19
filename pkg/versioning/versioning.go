package versioning

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func ParseVersion(version string) (*Version, error) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid version format")
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, err
	}

	return &Version{Major: major, Minor: minor, Patch: patch}, nil
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v *Version) IncrementMajor() {
	v.Major++
	v.Minor = 0
	v.Patch = 0
}

func (v *Version) IncrementMinor() {
	v.Minor++
	v.Patch = 0
}

func (v *Version) IncrementPatch() {
	v.Patch++
}