package envinfo

import (
	"reflect"
	"testing"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/kylelemons/godebug/pretty"
	"github.com/redhat-developer/odo/pkg/localConfigProvider"
	odoTestingUtil "github.com/redhat-developer/odo/pkg/testingutil"
)

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
