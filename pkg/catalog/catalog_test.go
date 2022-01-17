package catalog

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/redhat-developer/odo/pkg/preference"
)

func TestGetDevfileRegistries(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal("Fail to create temporary config file")
	}
	defer os.Remove(tempConfigFile.Name())
	defer tempConfigFile.Close()
	_, err = tempConfigFile.Write([]byte(
		`kind: Preference
apiversion: odo.openshift.io/v1alpha1
OdoSettings:
  RegistryList:
  - Name: DefaultDevfileRegistry
    URL: https://registry.devfile.io
  - Name: CheDevfileRegistry
    URL: https://che-devfile-registry.openshift.io/`,
	))
	if err != nil {
		t.Error(err)
	}

	os.Setenv(preference.GlobalConfigEnvName, tempConfigFile.Name())
	defer os.Unsetenv(preference.GlobalConfigEnvName)

	tests := []struct {
		name         string
		registryName string
		want         []Registry
	}{
		{
			name:         "Case 1: Test get all devfile registries",
			registryName: "",
			want: []Registry{
				{
					Name:   "CheDevfileRegistry",
					URL:    "https://che-devfile-registry.openshift.io/",
					Secure: false,
				},
				{
					Name:   "DefaultDevfileRegistry",
					URL:    "https://registry.devfile.io",
					Secure: false,
				},
			},
		},
		{
			name:         "Case 2: Test get specific devfile registry",
			registryName: "CheDevfileRegistry",
			want: []Registry{
				{
					Name:   "CheDevfileRegistry",
					URL:    "https://che-devfile-registry.openshift.io/",
					Secure: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDevfileRegistries(tt.registryName)
			if err != nil {
				t.Errorf("Error message is %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got: %v, want: %v", got, tt.want)
			}
		})
	}
}

func TestGetRegistryDevfiles(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		_, err := rw.Write([]byte(
			`
			[
				{
					"name": "nodejs",
					"displayName": "NodeJS Angular Web Application",
					"description": "Stack for developing NodeJS Angular Web Application",
					"tags": [
						"NodeJS",
						"Angular",
						"Alpine"
					],
					"language": "nodejs",
					"icon": "/images/angular.svg",
					"globalMemoryLimit": "2686Mi",
					"links": {
						"self": "/devfiles/angular/devfile.yaml"
					}
				}
			]
			`,
		))
		if err != nil {
			t.Error(err)
		}
	}))
	// Close the server when test finishes
	defer server.Close()

	const registryName = "some registry"
	tests := []struct {
		name     string
		registry Registry
		want     []DevfileComponentType
	}{
		{
			name:     "Test NodeJS devfile index",
			registry: Registry{Name: registryName, URL: server.URL},
			want: []DevfileComponentType{
				{
					Name:        "nodejs",
					DisplayName: "NodeJS Angular Web Application",
					Description: "Stack for developing NodeJS Angular Web Application",
					Registry: Registry{
						Name: registryName,
						URL:  server.URL,
					},
					Link:     "/devfiles/angular/devfile.yaml",
					Language: "nodejs",
					Tags:     []string{"NodeJS", "Angular", "Alpine"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getRegistryDevfiles(tt.registry)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got: %v, want: %v", got, tt.want)
				t.Logf("Error message is: %v", err)
			}
		})
	}
}

func TestConvertURL(t *testing.T) {
	tests := []struct {
		name    string
		URL     string
		wantURL string
	}{
		{
			name:    "Case 1: GitHub regular URL without specifying branch",
			URL:     "https://github.com/GeekArthur/registry",
			wantURL: "https://raw.githubusercontent.com/GeekArthur/registry/master",
		},
		{
			name:    "Case 2: GitHub regular URL with master branch specified",
			URL:     "https://github.ibm.com/Jingfu-J-Wang/registry/tree/master",
			wantURL: "https://raw.github.ibm.com/Jingfu-J-Wang/registry/master",
		},
		{
			name:    "Case 3: GitHub regular URL with non-master branch specified",
			URL:     "https://github.com/elsony/devfile-registry/tree/johnmcollier-crw",
			wantURL: "https://raw.githubusercontent.com/elsony/devfile-registry/johnmcollier-crw",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, err := convertURL(tt.URL)
			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(gotURL, tt.wantURL) {
				t.Errorf("Got url: %s, want URL: %s", gotURL, tt.wantURL)
			}
		})
	}
}

func TestDevfileComponentTypeList_GetLanguages(t *testing.T) {
	type fields struct {
		Items []DevfileComponentType
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "no devfiles",
			want: []string{},
		},
		{
			name: "some devfiles",
			fields: fields{
				Items: []DevfileComponentType{
					{
						Name:        "devfile4",
						DisplayName: "first devfile for lang3",
						Registry: Registry{
							Name: "Registry1",
						},
						Language: "lang3",
					},
					{
						Name:        "devfile1",
						DisplayName: "first devfile for lang1",
						Registry: Registry{
							Name: "Registry2",
						},
						Language: "lang1",
					},
					{
						Name:        "devfile3",
						DisplayName: "another devfile for lang2",
						Registry: Registry{
							Name: "Registry1",
						},
						Language: "lang2",
					},
					{
						Name:        "devfile2",
						DisplayName: "second devfile for lang1",
						Registry: Registry{
							Name: "Registry1",
						},
						Language: "lang1",
					},
				},
			},
			want: []string{"lang1", "lang2", "lang3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &DevfileComponentTypeList{
				Items: tt.fields.Items,
			}
			if got := o.GetLanguages(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DevfileComponentTypeList.GetLanguages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDevfileComponentTypeList_GetProjectTypes(t *testing.T) {
	type fields struct {
		Items []DevfileComponentType
	}
	type args struct {
		language string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   TypesWithDetails
	}{
		{
			name: "No devfiles => no project types",
			want: TypesWithDetails{},
		},
		{
			name: "project types for lang1",
			fields: fields{
				Items: []DevfileComponentType{
					{
						Name:        "devfile4",
						DisplayName: "first devfile for lang3",
						Registry: Registry{
							Name: "Registry1",
						},
						Language: "lang3",
					},
					{
						Name:        "devfile1",
						DisplayName: "first devfile for lang1",
						Registry: Registry{
							Name: "Registry1",
						},
						Language: "lang1",
					},
					{
						Name:        "devfile1",
						DisplayName: "first devfile for lang1",
						Registry: Registry{
							Name: "Registry2",
						},
						Language: "lang1",
					},
					{
						Name:        "devfile3",
						DisplayName: "another devfile for lang2",
						Registry: Registry{
							Name: "Registry1",
						},
						Language: "lang2",
					},
					{
						Name:        "devfile2",
						DisplayName: "second devfile for lang1",
						Registry: Registry{
							Name: "Registry1",
						},
						Language: "lang1",
					},
				},
			},
			args: args{
				language: "lang1",
			},
			want: TypesWithDetails{
				"first devfile for lang1": []DevfileComponentType{
					{
						Name:        "devfile1",
						DisplayName: "first devfile for lang1",
						Language:    "lang1",
						Registry: Registry{
							Name: "Registry1",
						},
					},
					{
						Name:        "devfile1",
						DisplayName: "first devfile for lang1",
						Language:    "lang1",
						Registry: Registry{
							Name: "Registry2",
						},
					},
				},
				"second devfile for lang1": []DevfileComponentType{
					{
						Name:        "devfile2",
						DisplayName: "second devfile for lang1",
						Language:    "lang1",
						Registry: Registry{
							Name: "Registry1",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &DevfileComponentTypeList{
				Items: tt.fields.Items,
			}
			if got := o.GetProjectTypes(tt.args.language); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DevfileComponentTypeList.GetProjectTypes() = \n%+v, want \n%+v", got, tt.want)
			}
		})
	}
}

func TestTypesWithDetails_GetOrderedLabels(t *testing.T) {
	tests := []struct {
		name  string
		types TypesWithDetails
		want  []string
	}{
		{
			name: "some entries",
			types: TypesWithDetails{
				"second devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile2",
						Registry: Registry{
							Name: "Registry1",
						},
					},
				},
				"first devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry1",
						},
					},
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry2",
						},
					},
				},
			},
			want: []string{
				"first devfile for lang1 (devfile1, registry: Registry1)",
				"first devfile for lang1 (devfile1, registry: Registry2)",
				"second devfile for lang1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.types.GetOrderedLabels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TypesWithDetails.GetOrderedLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypesWithDetails_GetAtOrderedPosition(t *testing.T) {
	type args struct {
		pos int
	}
	tests := []struct {
		name    string
		types   TypesWithDetails
		args    args
		want    DevfileComponentType
		wantErr bool
	}{
		{
			name: "get a pos 0",
			types: TypesWithDetails{
				"second devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile2",
						Registry: Registry{
							Name: "Registry1",
						},
					},
				},
				"first devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry1",
						},
					},
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry2",
						},
					},
				},
			},
			args: args{
				pos: 0,
			},
			want: DevfileComponentType{
				Name: "devfile1",
				Registry: Registry{
					Name: "Registry1",
				},
			},
			wantErr: false,
		},
		{
			name: "get a pos 1",
			types: TypesWithDetails{
				"second devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile2",
						Registry: Registry{
							Name: "Registry1",
						},
					},
				},
				"first devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry1",
						},
					},
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry2",
						},
					},
				},
			},
			args: args{
				pos: 1,
			},
			want: DevfileComponentType{
				Name: "devfile1",
				Registry: Registry{
					Name: "Registry2",
				},
			},
			wantErr: false,
		},
		{
			name: "get a pos 2",
			types: TypesWithDetails{
				"second devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile2",
						Registry: Registry{
							Name: "Registry1",
						},
					},
				},
				"first devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry1",
						},
					},
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry2",
						},
					},
				},
			},
			args: args{
				pos: 2,
			},
			want: DevfileComponentType{
				Name: "devfile2",
				Registry: Registry{
					Name: "Registry1",
				},
			},
			wantErr: false,
		},
		{
			name: "get a pos 4: not found",
			types: TypesWithDetails{
				"second devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile2",
						Registry: Registry{
							Name: "Registry1",
						},
					},
				},
				"first devfile for lang1": []DevfileComponentType{
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry1",
						},
					},
					{
						Name: "devfile1",
						Registry: Registry{
							Name: "Registry2",
						},
					},
				},
			},
			args: args{
				pos: 4,
			},
			want:    DevfileComponentType{},
			wantErr: true,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.types.GetAtOrderedPosition(tt.args.pos)
			if (err != nil) != tt.wantErr {
				t.Errorf("TypesWithDetails.GetAtOrderedPosition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TypesWithDetails.GetAtOrderedPosition() got1 = %v, want %v", got, tt.want)
			}
		})
	}
}
