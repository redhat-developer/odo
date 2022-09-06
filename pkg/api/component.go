package api

import (
	"sort"
	"strings"
)

type RunningMode string
type RunningModeList map[RunningMode]bool

const (
	RunningModeDev    RunningMode = "dev"
	RunningModeDeploy RunningMode = "deploy"
)

func NewRunningModeList() RunningModeList {
	return RunningModeList{
		RunningModeDev:    false,
		RunningModeDeploy: false,
	}
}

// AddRunningMode sets a running mode as true
// If mode is not "dev" or "deploy", an error is returned
func (o RunningModeList) AddRunningMode(mode RunningMode) {
	o[mode] = true
}

func (o RunningModeList) String() string {
	strs := make([]string, 0, len(o))
	for s, v := range o {
		if v {
			strs = append(strs, strings.Title(string(s)))
		}
	}
	if len(strs) == 0 {
		return "None"
	}
	sort.Sort(sort.Reverse(sort.StringSlice(strs)))
	return strings.Join(strs, ", ")
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
