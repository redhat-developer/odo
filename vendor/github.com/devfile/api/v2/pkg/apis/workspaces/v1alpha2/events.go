package v1alpha2

type Events struct {
	WorkspaceEvents `json:",inline"`
}

type WorkspaceEvents struct {
	// IDs of commands that should be executed before the workspace start.
	// Kubernetes-wise, these commands would typically be executed in init containers of the workspace POD.
	// +optional
	PreStart []string `json:"preStart,omitempty"`

	// IDs of commands that should be executed after the workspace is completely started.
	// In the case of Che-Theia, these commands should be executed after all plugins and extensions have started, including project cloning.
	// This means that those commands are not triggered until the user opens the IDE in his browser.
	// +optional
	PostStart []string `json:"postStart,omitempty"`

	// +optional
	// IDs of commands that should be executed before stopping the workspace.
	PreStop []string `json:"preStop,omitempty"`

	// +optional
	// IDs of commands that should be executed after stopping the workspace.
	PostStop []string `json:"postStop,omitempty"`
}
