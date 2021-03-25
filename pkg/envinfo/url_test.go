package envinfo

import (
	"github.com/devfile/library/pkg/devfile/parser/data"
	"reflect"
	"testing"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/kylelemons/godebug/pretty"
	"github.com/openshift/odo/pkg/localConfigProvider"
	odoTestingUtil "github.com/openshift/odo/pkg/testingutil"
)

func TestEnvInfo_CompleteURL(t *testing.T) {
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
		wantedURL localConfigProvider.LocalURL
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
					Container: "runtime-debug",
				},
			},
			wantedURL: localConfigProvider.LocalURL{
				Name:      "url-1",
				Port:      8080,
				Secure:    false,
				Path:      "/",
				Container: "runtime-debug",
			},
		},
		{
			name: "case 8: no container is present in the devfile and none is provided in the url",
			fields: fields{
				devfileObj: parser.DevfileObj{
					Data: func() data.DevfileData {
						devfileData, err := data.NewDevfileData(string(data.APIVersion200))
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
				Container: "runtime",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ei := &EnvInfo{
				devfileObj:        tt.fields.devfileObj,
				componentSettings: tt.fields.componentSettings,
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
		name    string
		fields  fields
		args    args
		wantErr bool
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
						devfileData, err := data.NewDevfileData(string(data.APIVersion200))
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ei := &EnvInfo{
				devfileObj:        tt.fields.devfileObj,
				componentSettings: tt.fields.componentSettings,
			}
			if err := ei.ValidateURL(tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnvInfo_GetPorts(t *testing.T) {
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

			got, err := ei.GetPorts()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPorts() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPorts() = %v, want %v", got, tt.want)
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
