package labels

import (
	"reflect"
	"testing"

	"github.com/openshift/odo/pkg/version"
)

func TestGetLabels(t *testing.T) {
	type args struct {
		applicationName string
		additional      bool
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Case 1: All labels",
			args: args{
				applicationName: "applicationame",
				additional:      false,
			},
			want: map[string]string{
				ApplicationLabel: "applicationame",
			},
		},
		{
			name: "Case 2: All labels including all additional labels",
			args: args{

				applicationName: "applicationame",
				additional:      true,
			},
			want: map[string]string{
				ApplicationLabel: "applicationame",
				App:              "applicationame",
				OdoManagedBy:     "odo",
				OdoVersion:       version.VERSION,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLabels(tt.args.applicationName, tt.args.additional); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
