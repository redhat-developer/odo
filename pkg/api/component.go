package api

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
