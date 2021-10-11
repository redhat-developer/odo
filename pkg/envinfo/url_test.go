package envinfo

import (
	"reflect"
	"testing"

	"github.com/devfile/library/pkg/devfile/parser/data"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/kylelemons/godebug/pretty"
	"github.com/openshift/odo/pkg/localConfigProvider"
	odoTestingUtil "github.com/openshift/odo/pkg/testingutil"
	"github.com/openshift/odo/pkg/util"
)

func TestEnvInfo_CompleteURL(t *testing.T) {
	fs := filesystem.NewFakeFs()

	type fields struct {
		devfileObj        parser.DevfileObj
		componentSettings ComponentSettings
		isRouteSupported  bool
	}
	type args struct {
		url localConfigProvider.LocalURL
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantedURL localConfigProvider.LocalURL
		updateURL bool
		wantErr   bool
	}{
		{
			name: "case 1: remove \\ from a path with length > 0 and complete the container with the first component container",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name: "url-1",
					Path: "\\data",
				},
			},
			wantedURL: localConfigProvider.LocalURL{
				Name:      "url-1",
				Port:      0,
				Secure:    false,
				Path:      "/data",
				Kind:      localConfigProvider.INGRESS,
				Container: "runtime",
			},
		},
		{
			name: "case 2: remove \\ from the path",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name: "url-1",
					Path: "\\",
				},
			},
			wantedURL: localConfigProvider.LocalURL{
				Name:      "url-1",
				Port:      0,
				Secure:    false,
				Path:      "/",
				Kind:      localConfigProvider.INGRESS,
				Container: "runtime",
			},
		},
		{
			name: "case 3: use the given path",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name: "url-1",
					Path: "/data",
				},
			},
			wantedURL: localConfigProvider.LocalURL{
				Name:      "url-1",
				Port:      0,
				Secure:    false,
				Path:      "/data",
				Kind:      localConfigProvider.INGRESS,
				Container: "runtime",
			},
		},
		{
			name: "case 4: complete the path when none is provided",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name: "url-1",
				},
			},
			wantedURL: localConfigProvider.LocalURL{
				Name:      "url-1",
				Port:      0,
				Secure:    false,
				Path:      "/",
				Kind:      localConfigProvider.INGRESS,
				Container: "runtime",
			},
		},
		{
			name: "case 5: complete the port when -1 is provided",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name: "url-1",
					Path: "\\data",
					Port: -1,
				},
			},
			wantedURL: localConfigProvider.LocalURL{
				Name:      "url-1",
				Port:      3000,
				Secure:    false,
				Path:      "/data",
				Kind:      localConfigProvider.INGRESS,
				Container: "runtime",
			},
		},
		{
			name: "case 6: complete the container based on the matching port in the devfile",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObjWithMultipleEndpoints(fs),
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name: "url-1",
					Port: 8080,
				},
			},
			wantedURL: localConfigProvider.LocalURL{
				Name:      "url-1",
				Port:      8080,
				Secure:    false,
				Path:      "/",
				Kind:      localConfigProvider.INGRESS,
				Container: "runtime-debug",
			},
		},
		{
			name: "case 7: do not change the container name given",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name:      "url-1",
					Port:      8080,
					Container: "runtime",
				},
			},
			wantedURL: localConfigProvider.LocalURL{
				Name:      "url-1",
				Port:      8080,
				Secure:    false,
				Path:      "/",
				Kind:      localConfigProvider.INGRESS,
				Container: "runtime",
			},
		},
		{
			name: "case 8: no container is present in the devfile and none is provided in the url",
			fields: fields{
				devfileObj: parser.DevfileObj{
					Data: func() data.DevfileData {
						devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
						if err != nil {
							t.Error(err)
						}
						return devfileData
					}(),
				},
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name: "url-1",
					Port: 8080,
				},
			},
			wantedURL: localConfigProvider.LocalURL{
				Name:   "url-1",
				Port:   8080,
				Secure: false,
				Kind:   localConfigProvider.INGRESS,
				Path:   "/",
			},
			wantErr: true,
		},
		{
			name: "case 9: complete the url name if not provided",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
				componentSettings: ComponentSettings{
					Name: "nodejs",
				},
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Port: 8080,
				},
			},
			wantedURL: localConfigProvider.LocalURL{
				Name:      "nodejs-8080",
				Port:      8080,
				Secure:    false,
				Path:      "/",
				Kind:      localConfigProvider.INGRESS,
				Container: "runtime",
			},
		},
		{
			name: "case 10: user doesn't provide an port and no ports are exposed by the devfile",
			fields: fields{
				devfileObj: parser.DevfileObj{
					Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
					Data: func() data.DevfileData {
						devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
						if err != nil {
							t.Error(err)
						}
						err = devfileData.AddComponents([]v1.Component{
							odoTestingUtil.GetFakeContainerComponent("runtime"),
						})
						if err != nil {
							t.Error(err)
						}
						return devfileData
					}(),
				},
				componentSettings: ComponentSettings{
					Name: "nodejs",
				},
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Port: -1,
				},
			},
			wantedURL: localConfigProvider.LocalURL{},
			wantErr:   true,
		},
		{
			name: "case 11: user doesn't provide an port and multiple ports are exposed by the devfile",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObjWithMultipleEndpoints(fs),
				componentSettings: ComponentSettings{
					Name: "nodejs",
				},
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Port: -1,
				},
			},
			wantedURL: localConfigProvider.LocalURL{},
			wantErr:   true,
		},
		{
			name: "case 12: complete the url kind if not provided and route is supported",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
				componentSettings: ComponentSettings{
					Name: "nodejs",
				},
				isRouteSupported: true,
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Port: 8080,
				},
			},
			wantedURL: localConfigProvider.LocalURL{
				Name:      "nodejs-8080",
				Port:      8080,
				Secure:    false,
				Path:      "/",
				Kind:      localConfigProvider.ROUTE,
				Container: "runtime",
			},
		},
		{
			name: "case 13: use an existing url when an invalid URL exists and no name and port is provided",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
				componentSettings: ComponentSettings{
					Name: "nodejs",
				},
				isRouteSupported: false,
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Port: -1,
				},
			},
			updateURL: true,
			wantedURL: localConfigProvider.LocalURL{
				Name:      "port-3030",
				Port:      3000,
				Secure:    false,
				Path:      "/",
				Kind:      localConfigProvider.INGRESS,
				Container: "runtime",
			},
		},
		{
			name: "case 14: use an existing url when an invalid URL exists and no name is provided but port is provided",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObjWithMultipleEndpoints(fs),
				componentSettings: ComponentSettings{
					Name: "nodejs",
				},
				isRouteSupported: false,
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Port: 3030,
				},
			},
			updateURL: true,
			wantedURL: localConfigProvider.LocalURL{
				Name:      "port-3030",
				Port:      3030,
				Secure:    false,
				Path:      "/",
				Kind:      localConfigProvider.INGRESS,
				Container: "runtime",
			},
		},
		{
			name: "case 15: Does not error out if no port is specified, but container with single port is specified in multi container devfile",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObjWithMultipleEndpoints(fs),
				componentSettings: ComponentSettings{
					Name: "nodejs",
				},
				isRouteSupported: false,
			},
			args: args{url: localConfigProvider.LocalURL{
				Secure:    false,
				Port:      -1,
				Container: "runtime-debug",
			}},
			wantErr:   false,
			updateURL: true,
			wantedURL: localConfigProvider.LocalURL{
				Name:      "port-8080",
				Port:      8080,
				Secure:    false,
				Path:      "/",
				Kind:      localConfigProvider.INGRESS,
				Container: "runtime-debug",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ei := &EnvInfo{
				devfileObj:        tt.fields.devfileObj,
				componentSettings: tt.fields.componentSettings,
				isRouteSupported:  tt.fields.isRouteSupported,
			}

			err := ei.CompleteURL(&tt.args.url)
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if tt.wantErr && err == nil {
				t.Errorf("wanted error but got no error")
			}

			if tt.wantErr && err != nil {
				return
			}
			if !reflect.DeepEqual(tt.args.url, tt.wantedURL) {
				t.Errorf("url doesn't match the required url: %v", pretty.Compare(tt.args.url, tt.wantedURL))
			}

			if tt.updateURL != ei.updateURL {
				t.Errorf("url update property doesn't match the required: %v", tt.updateURL)
			}
		})
	}
}

func TestEnvInfo_ValidateURL(t *testing.T) {
	fs := filesystem.NewFakeFs()

	type fields struct {
		devfileObj        parser.DevfileObj
		componentSettings ComponentSettings
	}
	type args struct {
		url localConfigProvider.LocalURL
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		updateURL bool
		wantErr   bool
	}{
		{
			name: "case 1: container not found",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name:      "runtime",
					Container: "blah",
				},
			},
			wantErr: true,
		},
		{
			name: "case 2: port occupied by another container",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name:      "runtime",
					Container: "loadbalancer",
					Port:      3000,
				},
			},
			wantErr: true,
		},

		{
			name: "case 3: tls secret used for non secure url",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name:      "runtime",
					TLSSecret: "blah",
				},
			},
			wantErr: true,
		},
		{
			name: "case 4: tls secret used for secure non ingress url",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name:      "runtime",
					TLSSecret: "blah",
					Secure:    true,
					Kind:      localConfigProvider.ROUTE,
				},
			},
			wantErr: true,
		},
		{
			name: "case 5: host used for Route based URL",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name: "runtime",
					Host: "com",
					Kind: localConfigProvider.ROUTE,
				},
			},
			wantErr: true,
		},
		{
			name: "case 6: host not provided for ingress URL",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name: "runtime",
					Kind: localConfigProvider.INGRESS,
				},
			},
			wantErr: true,
		},
		{
			name: "case 7: protocol not supported",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name:     "runtime",
					Protocol: "blah",
				},
			},
			wantErr: true,
		},
		{
			name: "case 8: url already exists",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
				componentSettings: ComponentSettings{
					URL: &[]localConfigProvider.LocalURL{
						{
							Name: "port-3030",
						},
					},
				},
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name: "port-3030",
				},
			},
			wantErr: true,
		},
		{
			name: "case 9: url already exists in a devfile endpoint",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
				componentSettings: ComponentSettings{
					URL: &[]localConfigProvider.LocalURL{},
				},
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name: "port-3030",
				},
			},
			wantErr: true,
		},
		{
			name: "case 10: no container found in devfile",
			fields: fields{
				devfileObj: parser.DevfileObj{
					Data: func() data.DevfileData {
						devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
						if err != nil {
							t.Error(err)
						}
						return devfileData
					}(),
				},
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name: "http-3000",
				},
			},
			wantErr: true,
		},
		{
			name: "case 11: host is not valid",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name: "runtime",
					Host: ",com",
					Kind: localConfigProvider.INGRESS,
				},
			},
			wantErr: true,
		},
		{
			name: "case 12: no error in the url",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name:      "http-3000",
					Secure:    true,
					TLSSecret: "blah",
					Host:      "com",
					Kind:      localConfigProvider.INGRESS,
				},
			},
			wantErr: false,
		},
		{
			name: "case 13: url exists but we are updating it",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
			},
			args: args{
				url: localConfigProvider.LocalURL{
					Name:      "port-3030",
					TLSSecret: "blah",
					Secure:    true,
					Host:      "com",
					Kind:      localConfigProvider.INGRESS,
				},
			},
			updateURL: true,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ei := &EnvInfo{
				devfileObj:        tt.fields.devfileObj,
				componentSettings: tt.fields.componentSettings,
				updateURL:         tt.updateURL,
			}
			if err := ei.ValidateURL(tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnvInfo_GetComponentPorts(t *testing.T) {
	fs := filesystem.NewFakeFs()

	type fields struct {
		devfileObj        parser.DevfileObj
		componentSettings ComponentSettings
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "case 1: multiple ports from multiple containers",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObjWithMultipleEndpoints(fs),
			},
			want: []string{"3000", "3030", "8080"},
		},
		{
			name: "case 2: single port from a container",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
			},
			want: []string{"3000"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ei := &EnvInfo{
				devfileObj:        tt.fields.devfileObj,
				componentSettings: tt.fields.componentSettings,
			}
			got, err := ei.GetComponentPorts()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetComponentPorts() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetComponentPorts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvInfo_GetContainerPorts(t *testing.T) {
	fs := filesystem.NewFakeFs()

	type fields struct {
		devfileObj        parser.DevfileObj
		componentSettings ComponentSettings
	}
	tests := []struct {
		name      string
		fields    fields
		container string
		want      []string
		wantErr   bool
	}{
		{
			name: "case 1: Returns ports of specified container in multi container devfile",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObjWithMultipleEndpoints(fs),
			},
			want:      []string{"8080"},
			container: "runtime-debug",
		},
		{
			name: "case 2: Returns error if no container is provided",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObjWithMultipleEndpoints(fs),
			},
			wantErr: true,
		},
		{
			name: "case 3: Returns error if invalid container is specified",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObjWithMultipleEndpoints(fs),
			},
			container: "invalidcontainer",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ei := &EnvInfo{
				devfileObj:        tt.fields.devfileObj,
				componentSettings: tt.fields.componentSettings,
			}
			got, err := ei.GetContainerPorts(tt.container)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetContainerPorts() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetContainerPorts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvInfo_ListURLs(t *testing.T) {
	fs := filesystem.NewFakeFs()

	type fields struct {
		devfileObj        parser.DevfileObj
		componentSettings ComponentSettings
	}
	tests := []struct {
		name    string
		fields  fields
		want    []localConfigProvider.LocalURL
		wantErr bool
	}{
		{
			name: "case 1: url present in devfile.yaml and env.yaml",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObjWithMultipleEndpoints(fs),
				componentSettings: ComponentSettings{
					URL: &[]localConfigProvider.LocalURL{
						{
							Name: "port-3030",
							Kind: localConfigProvider.INGRESS,
						},
						{
							Name: "port-3000",
							Kind: localConfigProvider.INGRESS,
						},
						{
							Name: "port-8080",
							Kind: localConfigProvider.INGRESS,
						},
					},
				},
			},
			want: []localConfigProvider.LocalURL{
				{
					Name:      "port-3030",
					Port:      3030,
					Container: "runtime",
					Path:      "/",
					Kind:      localConfigProvider.INGRESS,
				},
				{
					Name:      "port-3000",
					Port:      3000,
					Container: "runtime",
					Path:      "/",
					Kind:      localConfigProvider.INGRESS,
				},
				{
					Name:      "port-8080",
					Port:      8080,
					Container: "runtime-debug",
					Path:      "/",
					Kind:      localConfigProvider.INGRESS,
				},
			},
		},
		{
			name: "case 2: ignore URLs with none and internal endpoint",
			fields: fields{
				devfileObj: odoTestingUtil.DevfileObjWithInternalNoneEndpoints(fs),
				componentSettings: ComponentSettings{
					URL: &[]localConfigProvider.LocalURL{
						{
							Name: "port-3000",
							Kind: localConfigProvider.INGRESS,
						},
					},
				},
			},
			want: []localConfigProvider.LocalURL{
				{
					Name:      "port-3000",
					Port:      3000,
					Container: "runtime",
					Path:      "/",
					Kind:      localConfigProvider.INGRESS,
				},
			},
		},
		{
			name: "case 3: secure urls present in devfile.yaml with various protocols",
			fields: fields{
				devfileObj: odoTestingUtil.DevfileObjWithSecureEndpoints(fs),
				componentSettings: ComponentSettings{
					URL: &[]localConfigProvider.LocalURL{
						{
							Name: "port-3030",
							Kind: localConfigProvider.INGRESS,
						},
						{
							Name: "port-3000",
							Kind: localConfigProvider.INGRESS,
						},
						{
							Name: "port-8080",
							Kind: localConfigProvider.INGRESS,
						},
					},
				},
			},
			want: []localConfigProvider.LocalURL{
				{
					Name:      "port-3030",
					Port:      3030,
					Container: "runtime",
					Path:      "/",
					Secure:    true,
					Kind:      localConfigProvider.INGRESS,
				},
				{
					Name:      "port-3000",
					Port:      3000,
					Container: "runtime",
					Path:      "/",
					Secure:    true,
					Kind:      localConfigProvider.INGRESS,
				},
				{
					Name:      "port-8080",
					Port:      8080,
					Container: "runtime-debug",
					Path:      "/",
					Secure:    true,
					Kind:      localConfigProvider.INGRESS,
				},
			},
		},
		{
			name: "case 4: get the host, tlsSecret and kind from the env.yaml",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
				componentSettings: ComponentSettings{
					URL: &[]localConfigProvider.LocalURL{
						{
							Name:      "port-3030",
							Host:      "com",
							TLSSecret: "secret",
							Kind:      localConfigProvider.INGRESS,
						},
					},
				},
			},
			want: []localConfigProvider.LocalURL{
				{
					Name:      "port-3030",
					Port:      3000,
					Container: "runtime",
					Path:      "/",
					TLSSecret: "secret",
					Host:      "com",
					Kind:      localConfigProvider.INGRESS,
				},
			},
		},
		{
			name: "case 5: ignore the url present in the devfile.yaml but not in env.yaml",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
				componentSettings: ComponentSettings{
					URL: &[]localConfigProvider.LocalURL{
						{
							Name:      "port-3030",
							Host:      "com",
							TLSSecret: "secret",
							Kind:      localConfigProvider.INGRESS,
						},
						{
							Name: "port-8080",
							Host: "com",
							Kind: localConfigProvider.INGRESS,
						},
					},
				},
			},
			want: []localConfigProvider.LocalURL{
				{
					Name:      "port-3030",
					Port:      3000,
					Container: "runtime",
					Path:      "/",
					TLSSecret: "secret",
					Host:      "com",
					Kind:      localConfigProvider.INGRESS,
				},
			},
		},
		{
			name: "case 6: mark urls as route when present in devfile.yaml but not in env.yaml",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObj(fs),
				componentSettings: ComponentSettings{
					URL: &[]localConfigProvider.LocalURL{},
				},
			},
			want: []localConfigProvider.LocalURL{
				{
					Name:      "port-3030",
					Port:      3000,
					Container: "runtime",
					Path:      "/",
					Kind:      localConfigProvider.ROUTE,
				},
			},
		},
		{
			name: "case 7: use the path defined in the devfile.yaml",
			fields: fields{
				devfileObj: odoTestingUtil.GetTestDevfileObjWithPath(fs),
				componentSettings: ComponentSettings{
					URL: &[]localConfigProvider.LocalURL{
						{
							Name: "port-3030",
							Kind: localConfigProvider.INGRESS,
						},
					},
				},
			},
			want: []localConfigProvider.LocalURL{
				{
					Name:      "port-3030",
					Port:      3000,
					Container: "runtime",
					Path:      "/test",
					Kind:      localConfigProvider.INGRESS,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ei := &EnvInfo{
				devfileObj:        tt.fields.devfileObj,
				componentSettings: tt.fields.componentSettings,
			}

			got, err := ei.ListURLs()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListURLs() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListURLs() error: %v", pretty.Compare(got, tt.want))
			}
		})
	}
}

func Test_findInvalidEndpoint(t *testing.T) {
	fs := filesystem.NewFakeFs()

	type args struct {
		ei   *EnvInfo
		port int
	}
	tests := []struct {
		name    string
		args    args
		want    localConfigProvider.LocalURL
		wantErr bool
	}{
		{
			name: "case 1: find an invalid URL when route resources are not available",
			args: args{
				ei: &EnvInfo{
					isRouteSupported: false,
					devfileObj:       odoTestingUtil.GetTestDevfileObjWithMultipleEndpoints(fs),
				},
				port: 3030,
			},
			want: localConfigProvider.LocalURL{
				Name:      "port-3030",
				Port:      3030,
				Secure:    false,
				Kind:      localConfigProvider.ROUTE,
				Path:      "/",
				Container: "runtime",
			},
			wantErr: false,
		},
		{
			name: "case 2: route resources are available",
			args: args{
				ei: &EnvInfo{
					isRouteSupported: true,
					devfileObj:       odoTestingUtil.GetTestDevfileObjWithMultipleEndpoints(fs),
				},
				port: 3030,
			},
			want:    localConfigProvider.LocalURL{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findInvalidEndpoint(tt.args.ei, tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("findInvalidEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findInvalidEndpoint() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_updateEndpointInDevfile(t *testing.T) {
	fs := filesystem.NewFakeFs()

	type args struct {
		devObj parser.DevfileObj
		url    localConfigProvider.LocalURL
	}
	tests := []struct {
		name         string
		args         args
		wantEndpoint v1.Endpoint
		wantErr      bool
	}{
		{
			name: "case 1: update the endpoint when protocol is different",
			args: args{
				devObj: odoTestingUtil.DevfileObjWithSecureEndpoints(fs),
				url: localConfigProvider.LocalURL{
					Name:      "port-3030",
					Port:      3030,
					Protocol:  string(v1.WSSEndpointProtocol),
					Container: "runtime",
				},
			},
			wantEndpoint: v1.Endpoint{
				Name:       "port-3030",
				TargetPort: 3030,
				Protocol:   v1.WSSEndpointProtocol,
				Exposure:   v1.PublicEndpointExposure,
				Secure:     util.GetBoolPtr(false),
			},
			wantErr: false,
		},
		{
			name: "case 2: update the endpoint when exposure is different",
			args: args{
				devObj: odoTestingUtil.DevfileObjWithInternalNoneEndpoints(fs),
				url: localConfigProvider.LocalURL{
					Name:      "port-3030",
					Port:      3030,
					Container: "runtime",
				},
			},
			wantEndpoint: v1.Endpoint{
				Name:       "port-3030",
				TargetPort: 3030,
				Exposure:   v1.PublicEndpointExposure,
				Secure:     util.GetBoolPtr(false),
			},
			wantErr: false,
		},
		{
			name: "case 3: update the endpoint when path is different",
			args: args{
				devObj: odoTestingUtil.GetTestDevfileObjWithPath(fs),
				url: localConfigProvider.LocalURL{
					Name:      "port-3030",
					Port:      3000,
					Path:      "/user",
					Container: "runtime",
				},
			},
			wantEndpoint: v1.Endpoint{
				Name:       "port-3030",
				TargetPort: 3000,
				Path:       "/user",
				Exposure:   v1.PublicEndpointExposure,
				Secure:     util.GetBoolPtr(false),
			},
			wantErr: false,
		},
		{
			name: "case 4: update the endpoint when secure is different",
			args: args{
				devObj: odoTestingUtil.GetTestDevfileObj(fs),
				url: localConfigProvider.LocalURL{
					Name:      "port-3030",
					Port:      3000,
					Secure:    true,
					Container: "runtime",
				},
			},
			wantEndpoint: v1.Endpoint{
				Name:       "port-3030",
				TargetPort: 3000,
				Secure:     util.GetBoolPtr(true),
				Exposure:   v1.PublicEndpointExposure,
			},
			wantErr: false,
		},
		{
			name: "case 5: avoid a write when values are default",
			args: args{
				devObj: odoTestingUtil.GetTestDevfileObj(fs),
				url: localConfigProvider.LocalURL{
					Name:      "port-3030",
					Port:      3000,
					Container: "runtime",
					Protocol:  string(v1.HTTPEndpointProtocol),
					Path:      "/",
				},
			},
			wantEndpoint: v1.Endpoint{
				Name:       "port-3030",
				TargetPort: 3000,
			},
			wantErr: false,
		},
		{
			name: "case 6: url not found",
			args: args{
				devObj: odoTestingUtil.GetTestDevfileObj(fs),
				url: localConfigProvider.LocalURL{
					Name:      "port-303",
					Port:      3000,
					Container: "runtime",
					Protocol:  string(v1.HTTPEndpointProtocol),
					Path:      "/",
				},
			},
			wantEndpoint: v1.Endpoint{
				Name:       "port-3030",
				TargetPort: 3000,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := updateEndpointInDevfile(tt.args.devObj, tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("updateEndpointInDevfile() error = %v, wantErr %v", err, tt.wantErr)
			}

			components, err := tt.args.devObj.Data.GetComponents(parsercommon.DevfileOptions{})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			for _, component := range components {
				if component.Container != nil {
					for _, endpoint := range component.Container.Endpoints {
						if endpoint.Name == tt.args.url.Name {
							// prevent write unless required
							if !reflect.DeepEqual(tt.wantEndpoint, endpoint) {
								t.Errorf("expected endpoint doesn't match got: %v", pretty.Compare(tt.wantEndpoint, endpoint))
							}
						}
					}
				}
			}
		})
	}
}
