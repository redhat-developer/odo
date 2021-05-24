package labels

import (
	"reflect"
	"testing"

	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/version"
)

func TestGetLabels(t *testing.T) {
	type args struct {
		urlName         string
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
				urlName:         "urlname",
				componentName:   "componentname",
				applicationName: "applicationame",
				additional:      false,
			},
			want: map[string]string{
				applabels.ApplicationLabel:     "applicationame",
				componentlabels.ComponentLabel: "componentname",
				URLLabel:                       "urlname",
			},
		}, {
			name: "Case 2: No URL name",
			args: args{
				urlName:         "",
				componentName:   "componentname",
				applicationName: "applicationame",
				additional:      false,
			},
			want: map[string]string{
				applabels.ApplicationLabel:     "applicationame",
				componentlabels.ComponentLabel: "componentname",
				URLLabel:                       "",
			},
		}, {
			name: "Case 3: Everything with additional",
			args: args{
				urlName:         "urlname",
				componentName:   "componentname",
				applicationName: "applicationame",
				additional:      true,
			},
			want: map[string]string{
				applabels.ApplicationLabel:     "applicationame",
				applabels.App:                  "applicationame",
				applabels.ManagedBy:            "odo",
				applabels.ManagerVersion:       version.VERSION,
				componentlabels.ComponentLabel: "componentname",
				URLLabel:                       "urlname",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLabels(tt.args.urlName, tt.args.componentName, tt.args.applicationName, tt.args.additional); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
