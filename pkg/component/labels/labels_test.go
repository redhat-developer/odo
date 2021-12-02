package labels

import (
	"reflect"
	"testing"

	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	"github.com/redhat-developer/odo/pkg/version"
)

func TestGetLabels(t *testing.T) {
	type args struct {
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
				componentName:   "componentname",
				applicationName: "applicationame",
				additional:      false,
			},
			want: map[string]string{
				applabels.ApplicationLabel: "applicationame",
				ComponentLabel:             "componentname",
			},
		}, {
			name: "everything with additional",
			args: args{
				componentName:   "componentname",
				applicationName: "applicationame",
				additional:      true,
			},
			want: map[string]string{
				applabels.ApplicationLabel: "applicationame",
				applabels.App:              "applicationame",
				applabels.ManagedBy:        "odo",
				applabels.ManagerVersion:   version.VERSION,
				ComponentLabel:             "componentname",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLabels(tt.args.componentName, tt.args.applicationName, tt.args.additional); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
