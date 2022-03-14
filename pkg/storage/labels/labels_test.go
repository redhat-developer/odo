package labels

import (
	"reflect"
	"testing"

	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/version"
)

func TestGetLabels(t *testing.T) {
	type args struct {
		storageName     string
		componentName   string
		applicationName string
		additional      bool
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Case 1: Everything filled",
			args: args{
				storageName:     "storagename",
				componentName:   "componentname",
				applicationName: "applicationame",
				additional:      false,
			},
			want: map[string]string{
				applabels.ApplicationLabel:                       "applicationame",
				componentlabels.ComponentKubernetesInstanceLabel: "componentname",
				StorageLabel: "storagename",
			},
		}, {
			name: "Case 2: No storage name",
			args: args{
				storageName:     "",
				componentName:   "componentname",
				applicationName: "applicationame",
				additional:      false,
			},
			want: map[string]string{
				applabels.ApplicationLabel:                       "applicationame",
				componentlabels.ComponentKubernetesInstanceLabel: "componentname",
				StorageLabel: "",
			},
		}, {
			name: "Case 3: Everything with additional",
			args: args{
				storageName:     "storagename",
				componentName:   "componentname",
				applicationName: "applicationame",
				additional:      true,
			},
			want: map[string]string{
				applabels.ApplicationLabel:                       "applicationame",
				applabels.App:                                    "applicationame",
				applabels.ManagedBy:                              "odo",
				applabels.ManagerVersion:                         version.VERSION,
				componentlabels.ComponentKubernetesInstanceLabel: "componentname",
				StorageLabel:                                     "storagename",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLabels(tt.args.storageName, tt.args.componentName, tt.args.applicationName, tt.args.additional); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
