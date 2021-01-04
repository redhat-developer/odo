package util

import (
	"reflect"
	"testing"
)

func TestGetValidPortNumber(t *testing.T) {
	type args struct {
		portNumber    int
		componentName string
		portList      []string
	}
	tests := []struct {
		name       string
		args       args
		wantedPort int
		wantErr    bool
	}{
		{
			name: "test case 1: component with one port and port number provided",
			args: args{
				componentName: "nodejs",
				portNumber:    8080,
				portList:      []string{"8080/TCP"},
			},
			wantedPort: 8080,
			wantErr:    false,
		},
		{
			name: "test case 2: component with two ports and port number provided",
			args: args{
				componentName: "nodejs",
				portNumber:    8080,
				portList:      []string{"8080/TCP", "8081/TCP"},
			},
			wantedPort: 8080,
			wantErr:    false,
		},
		{
			name: "test case 3: service with two ports and no port number provided",
			args: args{
				componentName: "nodejs",
				portNumber:    -1,
				portList:      []string{"8080/TCP", "8081/TCP"},
			},

			wantErr: true,
		},
		{
			name: "test case 4: component with one port and no port number provided",
			args: args{
				componentName: "nodejs",
				portNumber:    -1,
				portList:      []string{"8080/TCP"},
			},
			wantedPort: 8080,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotPortNumber, err := GetValidPortNumber(tt.args.componentName, tt.args.portNumber, tt.args.portList)

			if err == nil && !tt.wantErr {

				if !reflect.DeepEqual(gotPortNumber, tt.wantedPort) {
					t.Errorf("Create() = %#v, want %#v", gotPortNumber, tt.wantedPort)
				}
			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			}
		})
	}
}
