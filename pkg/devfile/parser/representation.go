package parser

import (
	"github.com/openshift/odo/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d DevfileObj) ToRepresentation() DevfileComponentRepr {
	confRepr := DevfileComponentRepr{
		Name:   d.getMetadataName(),
		Memory: d.getMemory(),
	}
	var contReprs []ContainerRepr
	components := d.Data.GetComponents()
	for _, component := range components {

		if component.Container != nil {
			cont := ContainerRepr{
				ContainerName: component.Name,
			}
			cont.EnvironmentVariables = config.NewEnvVarListFromDevfileEnv(component.Container.Env)
			for _, endpoint := range component.Container.Endpoints {
				port := PortRepr{
					ExposedPort: endpoint.TargetPort,
					Name:        endpoint.Name,
					Protocol:    "http",
				}
				if endpoint.Protocol != "" {
					port.Protocol = string(endpoint.Protocol)
				}
				cont.Ports = append(cont.Ports, port)
			}
			contReprs = append(contReprs, cont)

		}
	}
	confRepr.Configs = contReprs
	return confRepr
}

func (d DevfileObj) WrapFromJSONOutput(confRepr DevfileComponentRepr) JSONConfigRepr {
	return JSONConfigRepr{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DevfileConfiguration",
			APIVersion: "odo.dev/v1alpha1",
		},
		DevfileConfigSpec: confRepr,
	}
}

type JSONConfigRepr struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	DevfileConfigSpec DevfileComponentRepr `json:"spec" yaml:"spec"`
}

type DevfileComponentRepr struct {
	Name    string          `yaml:"ComponentName,omitempty" json:"componentName,omitempty"`
	Memory  string          `yaml:"Memory,omitempty" json:"memory,omitempty"`
	Configs []ContainerRepr `yaml:"Configs,omitempty" json:"configs,omitempty"`

	// the parameter below are not configurables
	// Think of a better way
	State       string `yaml:"State,omitempty" json:"state,omitempty"`
	Namespace   string `yaml:"Namespace,omitempty" json:"namespace,omitempty"`
	Application string `yaml:"Application,omitempty" json:"application,omitempty"`
}

type ContainerRepr struct {
	ContainerName        string            `yaml:"ContainerName" json:"containerName"`
	EnvironmentVariables config.EnvVarList `yaml:"EnvironmentVariables" json:"environmentVariables,omitempty"`
	Ports                []PortRepr        `yaml:"Ports" json:"ports,omitempty"`
}

type PortRepr struct {
	Name        string `yaml:"Name" json:"name"`
	ExposedPort int32  `yaml:"ExposedPort" json:"exposedPort"`
	Protocol    string `yaml:"Protocol" json:"protocol"`
}
