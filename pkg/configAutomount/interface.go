package configAutomount

type MountAs int
type VolumeType int

const (
	MountAsFile MountAs = iota + 1
	MountAsSubpath
	MountAsEnv
)

const (
	VolumeTypePVC VolumeType = iota + 1
	VolumeTypeConfigmap
	VolumeTypeSecret
)

type AutomountInfo struct {
	// VolumeType gives the type of the volume (PVC, Secret, ConfigMap)
	VolumeType VolumeType
	// VolumeName is the name of the resource to mount
	VolumeName string
	// MountPath indicates on which path to mount the volume (empty if MountAs is Env)
	MountPath string
	// MountAs indicates how to mount the volume
	// - File: by default
	// - Env: As environment variables (for Secret and Configmap)
	// - Subpath: As individual files in specific paths (For Secret and ConfigMap). Keys must be provided
	MountAs MountAs
	// ReadOnly indicates to mount the volume as Read-Only
	ReadOnly bool
	// Keys defines the list of keys to mount when MountAs is Subpath
	Keys []string
}

type Client interface {
	GetAutomountingVolumes() ([]AutomountInfo, error)
}
