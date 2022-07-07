package watch

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
	State State
}

func componentCanSyncFile(state State) bool {
	return state == StateReady
}
