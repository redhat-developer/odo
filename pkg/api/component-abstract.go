package api

// ComponentAbstract represents a component as part of a list of components
type ComponentAbstract struct {
	Name      string          `json:"name"`
	ManagedBy string          `json:"managedBy"`
	RunningIn RunningModeList `json:"runningIn"`
	Type      string          `json:"projectType"`
}

const (
	// TypeUnknown means that odo cannot tell its state
	TypeUnknown = "Unknown"
	// TypeNone means that it has not been pushed to the cluster *at all* in either deploy or dev
	TypeNone = "None"
)
