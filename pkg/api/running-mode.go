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

// MergeRunningModes returns a new RunningModes map which is the result of merging all the running modes from each platform, from the map specified.
func MergeRunningModes(m map[string]RunningModes) RunningModes {
	if m == nil {
		return nil
	}

	rm := NewRunningModes()

	getMergedValueForMode := func(runningMode RunningMode) bool {
		for _, modeMap := range m {
			for mode, val := range modeMap {
				if mode != runningMode {
					continue
				}
				if val {
					return val
				}
			}
		}
		return false
	}

	// Dev
	if getMergedValueForMode(RunningModeDev) {
		rm.AddRunningMode(RunningModeDev)
	}

	// Deploy
	if getMergedValueForMode(RunningModeDeploy) {
		rm.AddRunningMode(RunningModeDeploy)
	}

	return rm
}
