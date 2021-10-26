package catalog

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/openshift/odo/v2/pkg/kclient"
	"github.com/openshift/odo/v2/pkg/preference"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestListOperatorServices(t *testing.T) {

	tests := []struct {
		name              string
		isCSVsupported    bool
		isCSVsupportedErr error
		list              *olm.ClusterServiceVersionList
		listErr           error
		expectedList      *olm.ClusterServiceVersionList
		expectedErr       bool
	}{
		{
			name:              "error getting supported csv",
			isCSVsupported:    false,
			isCSVsupportedErr: errors.New("an error"),
			expectedList:      &olm.ClusterServiceVersionList{},
			expectedErr:       true,
		},
		{
			name:              "non supported csv",
			isCSVsupported:    false,
			isCSVsupportedErr: nil,
			expectedList:      &olm.ClusterServiceVersionList{},
			expectedErr:       false,
		},
		{
			name:              "error getting list",
			isCSVsupported:    true,
			isCSVsupportedErr: nil,
			list:              nil,
			listErr:           errors.New("an error"),
			expectedList:      &olm.ClusterServiceVersionList{},
			expectedErr:       true,
		},
		{
			name:              "supported csv, empty list",
			isCSVsupported:    true,
			isCSVsupportedErr: nil,
			list:              &olm.ClusterServiceVersionList{},
			expectedList:      &olm.ClusterServiceVersionList{},
			expectedErr:       false,
		},
		{
			name:              "supported csv, return succeeded only",
			isCSVsupported:    true,
			isCSVsupportedErr: nil,
			list: &olm.ClusterServiceVersionList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "a kind",
					APIVersion: "a version",
				},
				Items: []olm.ClusterServiceVersion{
					{
						Status: olm.ClusterServiceVersionStatus{
							Phase: "Succeeded",
						},
					},
					{
						Status: olm.ClusterServiceVersionStatus{
							Phase: "",
						},
					},
					{
						Status: olm.ClusterServiceVersionStatus{
							Phase: "other phase",
						},
					},
				},
			},
			expectedList: &olm.ClusterServiceVersionList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "a kind",
					APIVersion: "a version",
				},
				Items: []olm.ClusterServiceVersion{
					{
						Status: olm.ClusterServiceVersionStatus{
							Phase: "Succeeded",
						},
					},
				},
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kc := kclient.NewMockClientInterface(ctrl)
			kc.EXPECT().IsCSVSupported().Return(tt.isCSVsupported, tt.isCSVsupportedErr).AnyTimes()
			kc.EXPECT().ListClusterServiceVersions().Return(tt.list, tt.listErr).AnyTimes()
			got, gotErr := ListOperatorServices(kc)
			if gotErr != nil != tt.expectedErr {
				t.Errorf("Got error %v, expected error %v\n", gotErr, tt.expectedErr)
			}
			if !reflect.DeepEqual(got, tt.expectedList) {
				t.Errorf("Got %v, expected %v\n", got, tt.expectedList)
			}
		})
	}
}
