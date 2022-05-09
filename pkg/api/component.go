package api

type RunningMode string
type RunningModeList []RunningMode

const (
	RunningModeDev     RunningMode = "Dev"
	RunningModeDeploy  RunningMode = "Deploy"
	RunningModeUnknown RunningMode = "Unknown"
)

func (u RunningModeList) Len() int {
	return len(u)
}
func (u RunningModeList) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}
func (u RunningModeList) Less(i, j int) bool {
	// Set Dev before Deploy
	return u[i] > u[j]
}

// Component describes the state of a devfile component
type Component struct {
	DevfilePath       string          `json:"devfilePath,omitempty"`
	DevfileData       *DevfileData    `json:"devfileData,omitempty"`
	DevForwardedPorts []ForwardedPort `json:"devForwardedPorts,omitempty"`
	RunningIn         []RunningMode   `json:"runningIn"`
	ManagedBy         string          `json:"managedBy"`
}

type ForwardedPort struct {
	ContainerName string `json:"containerName"`
	LocalAddress  string `json:"localAddress"`
	LocalPort     int    `json:"localPort"`
	ContainerPort int    `json:"containerPort"`
}
