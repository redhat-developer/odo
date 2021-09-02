package envinfo

import (
	"fmt"
	"io"
	"reflect"
	"text/tabwriter"

	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/machineoutput"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Info struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ComponentSettings `json:"spec"`
}

const InfoKind = "EnvInfo"

// ComponentSettings holds all component related information
type ComponentSettings struct {
	Name string `yaml:"Name,omitempty" json:"name,omitempty"`

	Project string `yaml:"Project,omitempty" json:"project,omitempty"`

	UserCreatedDevfile bool `yaml:"UserCreatedDevfile,omitempty" json:"UserCreatedDevfile,omitempty"`

	URL *[]localConfigProvider.LocalURL `yaml:"Url,omitempty" json:"url,omitempty"`
	// AppName is the application name. Application is a virtual concept present in odo used
	// for grouping of components. A namespace can contain multiple applications
	AppName string `yaml:"AppName,omitempty" json:"appName,omitempty"`

	// DebugPort controls the port used by the pod to run the debugging agent on
	DebugPort *int `yaml:"DebugPort,omitempty" json:"debugPort,omitempty"`

	// RunMode indicates the mode of run used for a successful push
	RunMode *RUNMode `yaml:"RunMode,omitempty" json:"runMode,omitempty"`
}

func NewInfo(cs ComponentSettings) Info {
	return Info{
		TypeMeta: metav1.TypeMeta{
			Kind:       InfoKind,
			APIVersion: machineoutput.APIVersion,
		},
		Spec: cs,
	}
}

func (o Info) Output(w io.Writer) {
	wr := tabwriter.NewWriter(w, 5, 2, 2, ' ', tabwriter.TabIndent)
	fmt.Fprintln(wr, "PARAMETER NAME", "\t", "PARAMETER VALUE")
	fmt.Fprintln(wr, "Name", "\t", o.Spec.Name)
	fmt.Fprintln(wr, "Project", "\t", o.Spec.Project)
	fmt.Fprintln(wr, "Application", "\t", o.Spec.AppName)
	fmt.Fprintln(wr, "DebugPort", "\t", showBlankIfNil(o.Spec.DebugPort))
	wr.Flush()
}

func showBlankIfNil(intf interface{}) interface{} {
	value := reflect.ValueOf(intf)

	// if the value is nil then we should return a blank string
	if value.IsNil() {
		return ""
	}

	// if it's a pointer then we should de-ref it because we cant de-ref an interface{}
	if value.Kind() == reflect.Ptr {
		return value.Elem().Interface()
	}

	return intf
}
