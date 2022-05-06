package registry

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
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
			prefClient, _ := preference.NewClient()
			catClient := NewRegistryClient(filesystem.NewFakeFs(), prefClient)
			got, err := catClient.GetDevfileRegistries(tt.registryName)
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
		want     []DevfileStack
	}{
		{
			name:     "Test NodeJS devfile index",
			registry: Registry{Name: registryName, URL: server.URL},
			want: []DevfileStack{
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
			ctrl := gomock.NewController(t)
			prefClient := preference.NewMockClient(ctrl)
			got, err := getRegistryStacks(prefClient, tt.registry)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got: %v, want: %v", got, tt.want)
				t.Logf("Error message is: %v", err)
			}
		})
	}
}
