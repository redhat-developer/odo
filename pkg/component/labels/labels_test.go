package labels

import (
	"reflect"
	"testing"

	"github.com/redhat-developer/odo/pkg/version"
	"k8s.io/apimachinery/pkg/labels"
)

func Test_getLabels(t *testing.T) {
	type args struct {
		componentName   string
		applicationName string
		additional      bool
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
				KubernetesManagedByLabel: "odo",
				KubernetesPartOfLabel:    "applicationame",
				KubernetesInstanceLabel:  "componentname",
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
				KubernetesPartOfLabel:           "applicationame",
				App:                             "applicationame",
				KubernetesManagedByLabel:        "odo",
				KubernetesManagedByVersionLabel: version.VERSION,
				KubernetesInstanceLabel:         "componentname",
				"odo.dev/mode":                  "Dev",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLabels(tt.args.componentName, tt.args.applicationName, ComponentDevMode, tt.args.additional); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
