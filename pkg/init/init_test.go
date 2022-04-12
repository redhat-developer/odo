package init

import (
	"errors"
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/alizer"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/registry"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func TestInitClient_downloadFromRegistry(t *testing.T) {
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
			o := &InitClient{
				preferenceClient: tt.fields.preferenceClient(ctrl),
				registryClient:   tt.fields.registryClient(ctrl),
			}
			if err := o.downloadFromRegistry(tt.args.registryName, tt.args.devfile, tt.args.dest); (err != nil) != tt.wantErr {
				t.Errorf("InitClient.downloadFromRegistry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInitClient_downloadDirect(t *testing.T) {
	type fields struct {
		fsys           func(fs filesystem.Filesystem) filesystem.Filesystem
		registryClient func(ctrl *gomock.Controller) registry.Client
		InitParams     alizer.DevfileLocation
	}
	type args struct {
		URL  string
		dest string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    func(fs filesystem.Filesystem) error
	}{
		{
			name: "download an existing file",
			fields: fields{
				fsys: func(fs filesystem.Filesystem) filesystem.Filesystem {
					_ = fs.WriteFile("/src/devfile.yaml", []byte("a content"), 0666)
					return fs
				},
				registryClient: func(ctrl *gomock.Controller) registry.Client {
					return nil
				},
			},
			args: args{
				URL:  "/src/devfile.yaml",
				dest: "/dest/file.yaml",
			},
			want: func(fs filesystem.Filesystem) error {
				content, err := fs.ReadFile("/dest/file.yaml")
				if err != nil {
					return errors.New("error reading file")
				}
				if string(content) != "a content" {
					return errors.New("content of file does not match")
				}
				info, err := fs.Stat("/dest/file.yaml")
				if err != nil {
					return errors.New("error executing Stat")
				}
				if info.Mode().Perm() != 0666 {
					return errors.New("permissions of destination file do not match")
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "non existing source file",
			fields: fields{
				fsys: func(fs filesystem.Filesystem) filesystem.Filesystem {
					return fs
				},
				registryClient: func(ctrl *gomock.Controller) registry.Client {
					return nil
				},
			},
			args: args{
				URL:  "/src/devfile.yaml",
				dest: "/dest/devfile.yaml",
			},
			want: func(fs filesystem.Filesystem) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "non existing URL",
			fields: fields{
				fsys: func(fs filesystem.Filesystem) filesystem.Filesystem {
					return fs
				},
				registryClient: func(ctrl *gomock.Controller) registry.Client {
					client := registry.NewMockClient(ctrl)
					client.EXPECT().DownloadFileInMemory(gomock.Any()).Return([]byte{}, errors.New(""))
					return client
				},
			},
			args: args{
				URL:  "https://example.com/devfile.yaml",
				dest: "/dest/devfile.yaml",
			},
			want: func(fs filesystem.Filesystem) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "existing URL",
			fields: fields{
				fsys: func(fs filesystem.Filesystem) filesystem.Filesystem {
					return fs
				},
				registryClient: func(ctrl *gomock.Controller) registry.Client {
					client := registry.NewMockClient(ctrl)
					client.EXPECT().DownloadFileInMemory(gomock.Any()).Return([]byte("a content"), nil)
					return client
				},
			},
			args: args{
				URL:  "https://example.com/devfile.yaml",
				dest: "/dest/devfile.yaml",
			},
			want: func(fs filesystem.Filesystem) error {
				content, err := fs.ReadFile("/dest/devfile.yaml")
				if err != nil {
					return errors.New("error reading dest file")
				}
				if string(content) != "a content" {
					return errors.New("unexpected file content")
				}
				return nil
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := filesystem.NewFakeFs()
			ctrl := gomock.NewController(t)
			o := &InitClient{
				fsys:           tt.fields.fsys(fs),
				registryClient: tt.fields.registryClient(ctrl),
			}
			if err := o.downloadDirect(tt.args.URL, tt.args.dest); (err != nil) != tt.wantErr {
				t.Errorf("InitClient.downloadDirect() error = %v, wantErr %v", err, tt.wantErr)
			}
			result := tt.want(fs)
			if result != nil {
				t.Errorf("unexpected error: %s", result)
			}
		})
	}
}

func TestInitClient_downloadStarterProject(t *testing.T) {
	type fields struct {
		registryClient func(ctrl *gomock.Controller) registry.Client
	}
	type args struct {
		project v1alpha2.StarterProject
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "starter project defined",
			fields: fields{
				registryClient: func(ctrl *gomock.Controller) registry.Client {
					client := registry.NewMockClient(ctrl)
					client.EXPECT().DownloadStarterProject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
					return client
				},
			},
			args: args{
				project: v1alpha2.StarterProject{
					Name: "project1",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			o := &InitClient{
				registryClient: tt.fields.registryClient(ctrl),
			}
			if err := o.DownloadStarterProject(&tt.args.project, "dest"); (err != nil) != tt.wantErr {
				t.Errorf("InitClient.downloadStarterProject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
