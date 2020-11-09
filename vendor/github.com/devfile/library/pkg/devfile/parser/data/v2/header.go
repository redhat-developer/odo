package v2

import (
	devfilepkg "github.com/devfile/api/pkg/devfile"
)

//SetSchemaVersion sets devfile schema version
func (d *DevfileV2) SetSchemaVersion(version string) {
	d.SchemaVersion = version
}

// GetMetadata returns the DevfileMetadata Object parsed from devfile
func (d *DevfileV2) GetMetadata() devfilepkg.DevfileMetadata {
	return d.Metadata
}

// SetMetadata sets the metadata for devfile
func (d *DevfileV2) SetMetadata(name, version string) {
	d.Metadata = devfilepkg.DevfileMetadata{
		Name:    name,
		Version: version,
	}
}
