package labels

import (
	"reflect"
	"testing"
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
			name: "everything",
			args: args{
				applicationName: "applicationame",
				additional:      false,
			},
			want: map[string]string{
				ApplicationLabel: "applicationame",
			},
		},
		{
			name: "everything with additional",
			args: args{

				applicationName: "applicationame",
				additional:      true,
			},
			want: map[string]string{
				ApplicationLabel:               "applicationame",
				AdditionalApplicationLabels[0]: "applicationame",
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
