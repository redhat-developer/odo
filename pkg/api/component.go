package api

import "strings"

type RunningMode string
type RunningModeList []RunningMode

const (
	RunningModeDev     RunningMode = "Dev"
	RunningModeDeploy  RunningMode = "Deploy"
	RunningModeUnknown RunningMode = "Unknown"
)

func (o RunningModeList) String() string {
	if len(o) == 0 {
		return "None"
	}
	strs := make([]string, 0, len(o))
	for _, s := range o {
		strs = append(strs, string(s))
	}
	return strings.Join(strs, ", ")
}

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
	RunningIn         RunningModeList `json:"runningIn"`
	ManagedBy         string          `json:"managedBy"`
}

type ForwardedPort struct {
	ContainerName string `json:"containerName"`
	LocalAddress  string `json:"localAddress"`
	LocalPort     int    `json:"localPort"`
	ContainerPort int    `json:"containerPort"`
}
