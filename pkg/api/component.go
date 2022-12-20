package api

import (
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type RunningMode string
type RunningModes map[RunningMode]bool

const (
	RunningModeDev    RunningMode = "dev"
	RunningModeDeploy RunningMode = "deploy"
)

func NewRunningModes() RunningModes {
	return RunningModes{
		RunningModeDev:    false,
		RunningModeDeploy: false,
	}
}

// AddRunningMode sets a running mode as true
func (o RunningModes) AddRunningMode(mode RunningMode) {
	o[mode] = true
}

func (o RunningModes) String() string {
	strs := make([]string, 0, len(o))
	caser := cases.Title(language.Und)
	for s, v := range o {
		if v {
			strs = append(strs, caser.String(string(s)))
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
	DevfilePath       string           `json:"devfilePath,omitempty"`
	DevfileData       *DevfileData     `json:"devfileData,omitempty"`
	DevForwardedPorts []ForwardedPort  `json:"devForwardedPorts,omitempty"`
	RunningIn         RunningModes     `json:"runningIn"`
	Ingresses         []ConnectionData `json:"ingresses,omitempty"`
	Routes            []ConnectionData `json:"routes,omitempty"`
	ManagedBy         string           `json:"managedBy"`
}

type ForwardedPort struct {
	Platform      string `json:"platform,omitempty"`
	ContainerName string `json:"containerName"`
	LocalAddress  string `json:"localAddress"`
	LocalPort     int    `json:"localPort"`
	ContainerPort int    `json:"containerPort"`
}

type ConnectionData struct {
	Name  string  `json:"name"`
	Rules []Rules `json:"rules,omitempty"`
}

type Rules struct {
	Host  string   `json:"host"`
	Paths []string `json:"paths"`
}
