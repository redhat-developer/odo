package init

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/odo/cli/init/params"
	"github.com/redhat-developer/odo/pkg/odo/cli/init/registry"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func TestInitOptions_Complete(t *testing.T) {
	type fields struct {
		backends func(*gomock.Controller) []params.ParamsBuilder
	}
	tests := []struct {
		name           string
		fields         fields
		cmdlineExpects func(*cmdline.MockCmdline)
		fsysPopulate   func(fsys filesystem.Filesystem)
		wantErr        bool
	}{
		{
			name: "directory not empty",
			fsysPopulate: func(fsys filesystem.Filesystem) {
				_ = fsys.WriteFile(".emptyfile", []byte(""), 0644)
			},
			wantErr: true,
		},
		{
			name: "second backend used",
			fields: fields{
				backends: func(ctrl *gomock.Controller) []params.ParamsBuilder {
					b1 := params.NewMockParamsBuilder(ctrl)
					b2 := params.NewMockParamsBuilder(ctrl)
					b1.EXPECT().IsAdequate(gomock.Any()).Return(false)
					b2.EXPECT().IsAdequate(gomock.Any()).Return(true)
					b2.EXPECT().ParamsBuild().Times(1)
					return []params.ParamsBuilder{b1, b2}
				},
			},
			cmdlineExpects: func(mock *cmdline.MockCmdline) {
				mock.EXPECT().GetFlags()
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := filesystem.NewFakeFs()
			if tt.fsysPopulate != nil {
				tt.fsysPopulate(fsys)
			}
			ctrl := gomock.NewController(t)
			var backends []params.ParamsBuilder
			if tt.fields.backends != nil {
				backends = tt.fields.backends(ctrl)
			}
			prefClient := preference.NewMockClient(ctrl)
			regClient := registry.NewMockClient(ctrl)
			o := NewInitOptions(backends, fsys, prefClient, regClient)

			cmdline := cmdline.NewMockCmdline(ctrl)
			if tt.cmdlineExpects != nil {
				tt.cmdlineExpects(cmdline)
			}
			if err := o.Complete(cmdline, []string{}); (err != nil) != tt.wantErr {
				t.Errorf("InitOptions.Complete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInitOptions_downloadFromRegistry(t *testing.T) {
	type fields struct {
		preferenceClient func(ctrl *gomock.Controller) preference.Client
		registryClient   func(ctrl *gomock.Controller) registry.Client
	}
	type args struct {
		registryName string
		devfile      string
		dest         string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Download devfile from one specific Registry where devfile is present",
			fields: fields{
				preferenceClient: func(ctrl *gomock.Controller) preference.Client {
					client := preference.NewMockClient(ctrl)
					registryList := []preference.Registry{
						{
							Name: "Registry0",
							URL:  "http://registry0",
						},
						{
							Name: "Registry1",
							URL:  "http://registry1",
						},
					}
					client.EXPECT().RegistryList().Return(&registryList)
					return client
				},
				registryClient: func(ctrl *gomock.Controller) registry.Client {
					client := registry.NewMockClient(ctrl)
					client.EXPECT().PullStackFromRegistry("http://registry1", "java", gomock.Any(), gomock.Any()).Return(nil).Times(1)
					return client
				},
			},
			args: args{
				registryName: "Registry1",
				devfile:      "java",
				dest:         ".",
			},
			wantErr: false,
		},
		{
			name: "Fail to download devfile from one specific Registry where devfile is absent",
			fields: fields{
				preferenceClient: func(ctrl *gomock.Controller) preference.Client {
					client := preference.NewMockClient(ctrl)
					registryList := []preference.Registry{
						{
							Name: "Registry0",
							URL:  "http://registry0",
						},
						{
							Name: "Registry1",
							URL:  "http://registry1",
						},
					}
					client.EXPECT().RegistryList().Return(&registryList)
					return client
				},
				registryClient: func(ctrl *gomock.Controller) registry.Client {
					client := registry.NewMockClient(ctrl)
					client.EXPECT().PullStackFromRegistry("http://registry1", "java", gomock.Any(), gomock.Any()).Return(errors.New("")).Times(1)
					return client
				},
			},
			args: args{
				registryName: "Registry1",
				devfile:      "java",
				dest:         ".",
			},
			wantErr: true,
		},
		{
			name: "Download devfile from all registries where devfile is present in second registry",
			fields: fields{
				preferenceClient: func(ctrl *gomock.Controller) preference.Client {
					client := preference.NewMockClient(ctrl)
					registryList := []preference.Registry{
						{
							Name: "Registry0",
							URL:  "http://registry0",
						},
						{
							Name: "Registry1",
							URL:  "http://registry1",
						},
					}
					client.EXPECT().RegistryList().Return(&registryList)
					return client
				},
				registryClient: func(ctrl *gomock.Controller) registry.Client {
					client := registry.NewMockClient(ctrl)
					client.EXPECT().PullStackFromRegistry("http://registry0", "java", gomock.Any(), gomock.Any()).Return(errors.New("")).Times(1)
					client.EXPECT().PullStackFromRegistry("http://registry1", "java", gomock.Any(), gomock.Any()).Return(nil).Times(1)
					return client
				},
			},
			args: args{
				registryName: "",
				devfile:      "java",
				dest:         ".",
			},
			wantErr: false,
		},
		{
			name: "Fail to download devfile from all registries where devfile is absent in all registries",
			fields: fields{
				preferenceClient: func(ctrl *gomock.Controller) preference.Client {
					client := preference.NewMockClient(ctrl)
					registryList := []preference.Registry{
						{
							Name: "Registry0",
							URL:  "http://registry0",
						},
						{
							Name: "Registry1",
							URL:  "http://registry1",
						},
					}
					client.EXPECT().RegistryList().Return(&registryList)
					return client
				},
				registryClient: func(ctrl *gomock.Controller) registry.Client {
					client := registry.NewMockClient(ctrl)
					client.EXPECT().PullStackFromRegistry("http://registry0", "java", gomock.Any(), gomock.Any()).Return(errors.New("")).Times(1)
					client.EXPECT().PullStackFromRegistry("http://registry1", "java", gomock.Any(), gomock.Any()).Return(errors.New("")).Times(1)
					return client
				},
			},
			args: args{
				registryName: "",
				devfile:      "java",
				dest:         ".",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			o := &InitOptions{
				preferenceClient: tt.fields.preferenceClient(ctrl),
				registryClient:   tt.fields.registryClient(ctrl),
			}
			if err := o.downloadFromRegistry(tt.args.registryName, tt.args.devfile, tt.args.dest); (err != nil) != tt.wantErr {
				t.Errorf("InitOptions.downloadFromRegistry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
