package api

type RunningMode string

const (
	RunningModeDev     RunningMode = "Dev"
	RunningModeDeploy  RunningMode = "Deploy"
	RunningModeUnknown RunningMode = "Unknown"
)

// Component describes the state of a devfile component
type Component struct {
	DevfilePath    string          `json:"devfilePath"`
	DevfileData    DevfileData     `json:"devfileData"`
	ForwardedPorts []ForwardedPort `json:"forwardedPorts"`
	RunningIn      []RunningMode   `json:"runningIn"`
	ManagedBy      string          `json:"managedBy"`
}

type ForwardedPort struct {
	ContainerName string `json:"containerName"`
	LocalPort     int    `json:"localPort"`
	ContainerPort int    `json:"containerPort"`
}
