package labels

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/redhat-developer/odo/pkg/version"
)

func Test_getLabels(t *testing.T) {
	type args struct {
		componentName     string
		applicationName   string
		additional        bool
		isPartOfComponent bool
	}
	tests := []struct {
		name string
		args args
		want labels.Set
	}{
		{
			name: "everything filled",
			args: args{
				componentName:   "componentname",
				applicationName: "applicationame",
				additional:      false,
			},
			want: labels.Set{
				kubernetesManagedByLabel: "odo",
				kubernetesPartOfLabel:    "applicationame",
				kubernetesInstanceLabel:  "componentname",
				"odo.dev/mode":           "Dev",
			},
		}, {
			name: "everything with additional",
			args: args{
				componentName:   "componentname",
				applicationName: "applicationame",
				additional:      true,
			},
			want: labels.Set{
				kubernetesPartOfLabel:           "applicationame",
				appLabel:                        "applicationame",
				kubernetesManagedByLabel:        "odo",
				kubernetesManagedByVersionLabel: version.VERSION,
				kubernetesInstanceLabel:         "componentname",
				"odo.dev/mode":                  "Dev",
			},
		},
		{
			name: "everything with isPartOfComponent",
			args: args{
				componentName:     "componentname",
				applicationName:   "applicationname",
				isPartOfComponent: true,
			},
			want: labels.Set{
				kubernetesManagedByLabel: "odo",
				kubernetesPartOfLabel:    "applicationname",
				kubernetesInstanceLabel:  "componentname",
				"odo.dev/mode":           "Dev",
				componentLabel:           "componentname",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLabels(tt.args.componentName, tt.args.applicationName, ComponentDevMode, tt.args.additional, tt.args.isPartOfComponent); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
