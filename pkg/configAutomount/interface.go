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
	VolumeType VolumeType
	VolumeName string
	MountPath  string
	MountAs    MountAs
	ReadOnly   bool
}

type Client interface {
	GetAutomountingVolumes() ([]AutomountInfo, error)
}
