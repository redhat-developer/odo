package parser

import (
	"github.com/openshift/odo/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d DevfileObj) ToRepresentation() ConfigurableRepr {
	confRepr := ConfigurableRepr{
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

func (d DevfileObj) WrapFromJSONOutput(confRepr ConfigurableRepr) JSONConfigRepr {
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
	DevfileConfigSpec ConfigurableRepr `json:"spec" yaml:"spec"`
}

type ConfigurableRepr struct {
	Name    string          `yaml:"ComponentName,omitempty" json:"ComponentName,omitempty"`
	Memory  string          `yaml:"Memory,omitempty" json:"Memory,omitempty"`
	Configs []ContainerRepr `yaml:"Configs,omitempty" json:"Configs,omitempty"`
}

type ContainerRepr struct {
	ContainerName        string            `yaml:"ContainerName" json:"ContainerName"`
	EnvironmentVariables config.EnvVarList `yaml:"EnvironmentVariables" json:"EnvironmentVariables,omitempty"`
	Ports                []PortRepr        `yaml:"Ports" json:"Ports,omitempty"`
}

type PortRepr struct {
	Name        string `yaml:"Name" json:"Name"`
	ExposedPort int32  `yaml:"ExposedPort" json:"ExposedPort"`
	Protocol    string `yaml:"Protocol" json:"Protocol"`
}
