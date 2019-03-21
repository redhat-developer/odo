package labels

import (
	"reflect"
	"testing"

	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
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
			name: "everything filled",
			args: args{
				storageName:     "storagename",
				componentName:   "componentname",
				applicationName: "applicationame",
				additional:      false,
			},
			want: map[string]string{
				applabels.ApplicationLabel:     "applicationame",
				componentlabels.ComponentLabel: "componentname",
				StorageLabel:                   "storagename",
			},
		}, {
			name: "no storage name",
			args: args{
				storageName:     "",
				componentName:   "componentname",
				applicationName: "applicationame",
				additional:      false,
			},
			want: map[string]string{
				applabels.ApplicationLabel:     "applicationame",
				componentlabels.ComponentLabel: "componentname",
				StorageLabel:                   "",
			},
		}, {
			name: "everything with additional",
			args: args{
				storageName:     "storagename",
				componentName:   "componentname",
				applicationName: "applicationame",
				additional:      true,
			},
			want: map[string]string{
				applabels.ApplicationLabel:               "applicationame",
				applabels.AdditionalApplicationLabels[0]: "applicationame",
				componentlabels.ComponentLabel:           "componentname",
				StorageLabel:                             "storagename",
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
