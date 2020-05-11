package api

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// invalidFilenameCharacters contains a list of character we consider malicious
// when injecting the directories into containers.
const invalidFilenameCharacters = `;*?"<>|%#$!+{}&[],"'` + "`"

// VolumeSpec represents a single volume mount point.
type VolumeSpec struct {
	// Source is a reference to the volume source.
	Source string
	// Destination is the path to mount the volume to - absolute or relative.
	Destination string
	// Keep indicates if the mounted data should be kept in the final image.
	Keep bool
}

// VolumeList contains list of VolumeSpec.
type VolumeList []VolumeSpec

// IsInvalidFilename verifies if the provided filename contains malicious
// characters.
func IsInvalidFilename(name string) bool {
	return strings.ContainsAny(name, invalidFilenameCharacters)
}

// Set implements the Set() function of pflags.Value interface.
// This function parses the string that contains source:destination pair.
// When the destination is not specified, the source get copied into current
// working directory in container.
func (l *VolumeList) Set(value string) error {
	volumes := strings.Split(value, ";")
	newVols := make([]VolumeSpec, len(volumes))
	for i, v := range volumes {
		spec, err := l.parseSpec(v)
		if err != nil {
			return err
		}
		newVols[i] = *spec
	}
	*l = append(*l, newVols...)
	return nil
}

func (l *VolumeList) parseSpec(value string) (*VolumeSpec, error) {
	if len(value) == 0 {
		return nil, errors.New("invalid format, must be source:destination")
	}
	var mount []string
	pos := strings.LastIndex(value, ":")
	if pos == -1 {
		mount = []string{value, ""}
	} else {
		mount = []string{value[:pos], value[pos+1:]}
	}
	mount[0] = strings.Trim(mount[0], `"'`)
	mount[1] = strings.Trim(mount[1], `"'`)
	s := &VolumeSpec{Source: filepath.Clean(mount[0]), Destination: filepath.ToSlash(filepath.Clean(mount[1]))}
	if IsInvalidFilename(s.Source) || IsInvalidFilename(s.Destination) {
		return nil, fmt.Errorf("invalid characters in filename: %q", value)
	}
	return s, nil
}

// String implements the String() function of pflags.Value interface.
func (l *VolumeList) String() string {
	result := []string{}
	for _, i := range *l {
		result = append(result, strings.Join([]string{i.Source, i.Destination}, ":"))
	}
	return strings.Join(result, ",")
}

// Type implements the Type() function of pflags.Value interface.
func (l *VolumeList) Type() string {
	return "string"
}
