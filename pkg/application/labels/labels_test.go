package labels

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/openshift/odo/v2/pkg/version"
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
				ManagedBy:        "odo",
				ManagerVersion:   version.VERSION,
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

func TestGetSelector(t *testing.T) {
	app := "sample-app"
	got := GetSelector(app)
	wants := []string{fmt.Sprintf("%v=%v", ApplicationLabel, app), fmt.Sprintf("%v=odo", ManagedBy)}
	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Errorf("got: %q, want: %q", got, want)
		}
	}
}

func TestGetNonOdoSelector(t *testing.T) {
	app := "sample-app"
	got := GetNonOdoSelector(app)
	wants := []string{fmt.Sprintf("%v=%v", ApplicationLabel, app), fmt.Sprintf("%v!=odo", ManagedBy)}
	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Errorf("got: %q, want: %q", got, want)
		}
	}
}
