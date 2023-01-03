package api

// ComponentAbstract represents a component as part of a list of components
type ComponentAbstract struct {
	Name             string `json:"name"`
	ManagedBy        string `json:"managedBy"`
	ManagedByVersion string `json:"managedByVersion"`
	// RunningIn are the modes the component is running in, among Dev and Deploy
	RunningIn RunningModes `json:"runningIn"`
	Type      string       `json:"projectType"`
	// RunningOn is the platform the component is running on, either cluster or podman
	//
	// Deprecated: This field is deprecated and will be replaced by Platform
	RunningOn string `json:"runningOn,omitempty"`
	// Platform is the platform the component is running on, either cluster or podman
	Platform string `json:"platform,omitempty"`
}

const (
	// TypeUnknown means that odo cannot tell its state
	TypeUnknown = "Unknown"
	// TypeNone means that it has not been pushed to the cluster *at all* in either deploy or dev
	TypeNone = "None"
)
