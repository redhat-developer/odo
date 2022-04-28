package state

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func TestState_SetForwardedPorts(t *testing.T) {

	forwardedPort1 := api.ForwardedPort{
		ContainerName: "acontainer",
		LocalAddress:  "localhost",
		LocalPort:     40001,
		ContainerPort: 3000,
	}

	type fields struct {
		fs                  func() filesystem.Filesystem
		getSecondsFromEpoch func() int64
		getpid              func() int
	}
	type args struct {
		fwPorts []api.ForwardedPort
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		checkState func(fs filesystem.Filesystem) error
	}{
		// TODO: Add test cases.
		{
			name: "set forwarded ports",
			fields: fields{
				fs: func() filesystem.Filesystem {
					return filesystem.NewFakeFs()
				},
				getSecondsFromEpoch: func() int64 {
					return 13000
				},
				getpid: func() int {
					return 100
				},
			},
			args: args{
				fwPorts: []api.ForwardedPort{forwardedPort1},
			},
			wantErr: false,
			checkState: func(fs filesystem.Filesystem) error {
				jsonContent, err := fs.ReadFile(_filepath)
				if err != nil {
					return err
				}
				var content Content
				err = json.Unmarshal(jsonContent, &content)
				if err != nil {
					return err
				}
				if content.Timestamp != 13000 {
					return fmt.Errorf("timestamp is %d, should be 13000", content.Timestamp)
				}
				if content.PID != 100 {
					return fmt.Errorf("PID is %d, should be 100", content.PID)
				}
				expected := []api.ForwardedPort{forwardedPort1}
				if !reflect.DeepEqual(content.ForwardedPorts, expected) {
					return fmt.Errorf("Forwarded ports is %+v, should be %+v", content.ForwardedPorts, expected)
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := tt.fields.fs()
			o := State{
				fs:                  fs,
				getSecondsFromEpoch: tt.fields.getSecondsFromEpoch,
				getpid:              tt.fields.getpid,
			}
			if err := o.SetForwardedPorts(tt.args.fwPorts); (err != nil) != tt.wantErr {
				t.Errorf("State.SetForwardedPorts() error = %v, wantErr %v", err, tt.wantErr)
			}
			if check := tt.checkState(fs); check != nil {
				t.Error(check)
			}
		})
	}
}

func TestState_SaveExit(t *testing.T) {
	type fields struct {
		fs                  func() filesystem.Filesystem
		getSecondsFromEpoch func() int64
		getpid              func() int
	}
	tests := []struct {
		name       string
		fields     fields
		wantErr    bool
		checkState func(fs filesystem.Filesystem) error
	}{
		{
			name: "save exit",
			fields: fields{
				fs: func() filesystem.Filesystem {
					return filesystem.NewFakeFs()
				},
				getSecondsFromEpoch: func() int64 {
					return 13000
				},
				getpid: func() int {
					return 100
				},
			},
			wantErr: false,
			checkState: func(fs filesystem.Filesystem) error {
				jsonContent, err := fs.ReadFile(_filepath)
				if err != nil {
					return err
				}
				var content Content
				err = json.Unmarshal(jsonContent, &content)
				if err != nil {
					return err
				}
				if content.Timestamp != 13000 {
					return fmt.Errorf("timestamp is %d, should be 13000", content.Timestamp)
				}
				if content.PID != 0 {
					return fmt.Errorf("PID is %d, should be 0", content.PID)
				}
				if len(content.ForwardedPorts) != 0 {
					return fmt.Errorf("Forwarded ports is %+v, should be empty", content.ForwardedPorts)
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := tt.fields.fs()
			o := State{
				fs:                  fs,
				getSecondsFromEpoch: tt.fields.getSecondsFromEpoch,
				getpid:              tt.fields.getpid,
			}
			if err := o.SaveExit(); (err != nil) != tt.wantErr {
				t.Errorf("State.SaveExit() error = %v, wantErr %v", err, tt.wantErr)
			}
			if check := tt.checkState(fs); check != nil {
				t.Error(check)
			}
		})
	}
}
