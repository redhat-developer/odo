package watch

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"k8s.io/klog"
)

type State string

const (
	StateWaitDeployment State = "WaitDeployment"
	StateSyncOutdated   State = "SyncOutdated"
	//StateWaitBindings         State = "WaitBindings"
	//StatePodRunning           State = "PodRunning"
	//StateFilesSynced          State = "FilesSynced"
	//StateBuildCommandExecuted State = "BuildCommandExecuted"
	//StateRunCommandRunning    State = "RunCommandRunning"
	StateReady State = "Ready"
)

type ComponentStatus struct {
	state               State
	PostStartEventsDone bool
	// RunExecuted is set to true when the run command has been executed
	// Used for HotReload capability
	RunExecuted        bool
	EndpointsForwarded map[string][]v1alpha2.Endpoint
	// ImageComponentsAutoApplied is a cache of all image components that have been auto-applied.
	// This map allows to avoid applying them too many times upon state changes in the cluster for example.
	ImageComponentsAutoApplied map[string]v1alpha2.ImageComponent
}

func (o *ComponentStatus) SetState(s State) {
	klog.V(4).Infof("setting inner loop State %q", s)
	o.state = s
}

func (o *ComponentStatus) GetState() State {
	return o.state
}

func componentCanSyncFile(state State) bool {
	return state == StateReady
}
