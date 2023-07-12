package api

// Component describes the state of a devfile component
type Component struct {
	DevfilePath       string          `json:"devfilePath,omitempty"`
	DevfileData       *DevfileData    `json:"devfileData,omitempty"`
	DevForwardedPorts []ForwardedPort `json:"devForwardedPorts,omitempty"`
	// RunningIn is the overall running mode map of the component;
	// this is computing as a merge of RunningOn (all the different running modes
	// for each platform the component is running on).
	RunningIn RunningModes `json:"runningIn"`
	// RunningOn represents the map of running modes for each platform the component is running on.
	// The key is the platform, either cluster or podman.
	RunningOn map[string]RunningModes `json:"runningOn,omitempty"`
	Ingresses []ConnectionData        `json:"ingresses,omitempty"`
	Routes    []ConnectionData        `json:"routes,omitempty"`
	ManagedBy string                  `json:"managedBy"`
}

type ForwardedPort struct {
	Platform      string `json:"platform,omitempty"`
	ContainerName string `json:"containerName"`
	PortName      string `json:"portName"`
	IsDebug       bool   `json:"isDebug"`
	LocalAddress  string `json:"localAddress"`
	LocalPort     int    `json:"localPort"`
	ContainerPort int    `json:"containerPort"`
	Exposure      string `json:"exposure,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
}

type DevControlPlane struct {
	Platform         string `json:"platform,omitempty"`
	LocalPort        int    `json:"localPort"`
	APIServerPath    string `json:"apiServerPath"`
	WebInterfacePath string `json:"webInterfacePath"`
}

type ConnectionData struct {
	Name  string  `json:"name"`
	Rules []Rules `json:"rules,omitempty"`
}

type Rules struct {
	Host  string   `json:"host"`
	Paths []string `json:"paths"`
}
