package state

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/redhat-developer/odo/pkg/api"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func TestState_SetForwardedPorts(t *testing.T) {

	forwardedPort1 := api.ForwardedPort{
		ContainerName: "acontainer",
		LocalAddress:  "localhost",
		LocalPort:     20001,
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
				expected := []api.ForwardedPort{forwardedPort1}
				if diff := cmp.Diff(expected, content.ForwardedPorts); diff != "" {
					return fmt.Errorf("forwarded ports is %+v, should be %+v, diff: %s", content.ForwardedPorts, expected, diff)
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := tt.fields.fs()
			o := State{
				fs: fs,
			}
			ctx := context.Background()
			ctx = odocontext.WithPID(ctx, 1)
			if err := o.SetForwardedPorts(ctx, tt.args.fwPorts); (err != nil) != tt.wantErr {
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
				fs: fs,
			}
			ctx := context.Background()
			ctx = odocontext.WithPID(ctx, 1)
			_ = o.SetForwardedPorts(ctx, nil)
			if err := o.SaveExit(ctx); (err != nil) != tt.wantErr {
				t.Errorf("State.SaveExit() error = %v, wantErr %v", err, tt.wantErr)
			}
			if check := tt.checkState(fs); check != nil {
				t.Error(check)
			}
		})
	}
}

func TestState_GetForwardedPorts(t *testing.T) {
	contentPodman := Content{
		Platform: "podman",
		ForwardedPorts: []api.ForwardedPort{
			{
				ContainerName: "acontainer",
				LocalAddress:  "localhost",
				LocalPort:     20001,
				ContainerPort: 3000,
			},
		},
	}
	contentCluster := Content{
		Platform: "cluster",
		ForwardedPorts: []api.ForwardedPort{
			{
				ContainerName: "acontainer",
				LocalAddress:  "localhost",
				LocalPort:     20002,
				ContainerPort: 3000,
			},
		},
	}
	type fields struct {
		content Content
		fs      func(t *testing.T) filesystem.Filesystem
	}
	tests := []struct {
		name    string
		fields  fields
		want    []api.ForwardedPort
		wantErr bool
	}{
		{
			name: "get forwarded ports, only deployed on podman",
			fields: fields{
				content: Content{},
				fs: func(t *testing.T) filesystem.Filesystem {
					fs := filesystem.NewFakeFs()
					jsonContent, err := json.Marshal(contentPodman)
					if err != nil {
						t.Errorf("Error marshaling data")
					}
					pid := 1
					err = fs.WriteFile(getFilename(pid), jsonContent, 0644)
					if err != nil {
						t.Errorf("Error saving content to file")
					}
					return fs
				},
			},
			want:    contentPodman.ForwardedPorts,
			wantErr: false,
		},
		{
			name: "get forwarded ports, only deployed on cluster",
			fields: fields{
				content: Content{},
				fs: func(t *testing.T) filesystem.Filesystem {
					fs := filesystem.NewFakeFs()
					jsonContent, err := json.Marshal(contentCluster)
					if err != nil {
						t.Errorf("Error marshaling data")
					}
					pid := 1
					err = fs.WriteFile(getFilename(pid), jsonContent, 0644)
					if err != nil {
						t.Errorf("Error saving content to file")
					}
					return fs
				},
			},
			want:    contentCluster.ForwardedPorts,
			wantErr: false,
		},
		{
			name: "get forwarded ports, deployed on both podman and cluster",
			fields: fields{
				content: Content{},
				fs: func(t *testing.T) filesystem.Filesystem {
					fs := filesystem.NewFakeFs()

					pidCluster := 1
					jsonContentCluster, err := json.Marshal(contentCluster)
					if err != nil {
						t.Errorf("Error marshaling data")
					}
					err = fs.WriteFile(getFilename(pidCluster), jsonContentCluster, 0644)
					if err != nil {
						t.Errorf("Error saving content to file")
					}

					pidPodman := 2
					jsonContentPodman, err := json.Marshal(contentPodman)
					if err != nil {
						t.Errorf("Error marshaling data")
					}
					err = fs.WriteFile(getFilename(pidPodman), jsonContentPodman, 0644)
					if err != nil {
						t.Errorf("Error saving content to file")
					}

					return fs
				},
			},
			want:    append(append([]api.ForwardedPort{}, contentCluster.ForwardedPorts...), contentPodman.ForwardedPorts...),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &State{
				content: tt.fields.content,
				fs:      tt.fields.fs(t),
			}
			ctx := context.Background()
			ctx = odocontext.WithPID(ctx, 1)
			got, err := o.GetForwardedPorts(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("State.GetForwardedPorts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("State.GetForwardedPorts() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
