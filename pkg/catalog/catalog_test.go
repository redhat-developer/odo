package catalog

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	imagev1 "github.com/openshift/api/image/v1"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/preference"
	"github.com/openshift/odo/pkg/testingutil"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestListComponents(t *testing.T) {
	type args struct {
		name       string
		namespace  string
		tags       []string
		hiddenTags []string
	}
	tests := []struct {
		name              string
		args              args
		wantErr           bool
		wantAllTags       []string
		wantNonHiddenTags []string
	}{
		{
			name: "Case 1: Valid image output with one tag which is not hidden",
			args: args{
				name:       "foobar",
				namespace:  "openshift",
				tags:       []string{"latest"},
				hiddenTags: []string{},
			},
			wantErr:           false,
			wantAllTags:       []string{"latest"},
			wantNonHiddenTags: []string{"latest"},
		},
		{
			name: "Case 2: Valid image output with one tag which is hidden",
			args: args{
				name:       "foobar",
				namespace:  "openshift",
				tags:       []string{"latest"},
				hiddenTags: []string{"latest"},
			},
			wantErr:           false,
			wantAllTags:       []string{"latest"},
			wantNonHiddenTags: []string{},
		},
		{
			name: "Case 3: Valid image output with multiple tags none of which are hidden",
			args: args{
				name:       "foobar",
				namespace:  "openshift",
				tags:       []string{"1.0.0", "1.0.1", "0.0.1", "latest"},
				hiddenTags: []string{},
			},
			wantErr:           false,
			wantAllTags:       []string{"1.0.0", "1.0.1", "0.0.1", "latest"},
			wantNonHiddenTags: []string{"1.0.0", "1.0.1", "0.0.1", "latest"},
		},
		{
			name: "Case 4: Valid image output with multiple tags some of which are hidden",
			args: args{
				name:       "foobar",
				namespace:  "openshift",
				tags:       []string{"1.0.0", "1.0.1", "0.0.1", "latest"},
				hiddenTags: []string{"0.0.1", "1.0.0"},
			},
			wantErr:           false,
			wantAllTags:       []string{"1.0.0", "1.0.1", "0.0.1", "latest"},
			wantNonHiddenTags: []string{"1.0.1", "latest"},
		},
		{
			name: "Case 3: Invalid image output with no tags",
			args: args{
				name:      "foobar",
				namespace: "foo",
				tags:      []string{},
			},
			wantErr:           true,
			wantAllTags:       []string{},
			wantNonHiddenTags: []string{},
		},
		{
			name: "Case 4: Valid image with output tags from a different namespace none of which are hidden",
			args: args{
				name:       "foobar",
				namespace:  "foo",
				tags:       []string{"1", "2", "4", "latest", "10"},
				hiddenTags: []string{"1", "2"},
			},
			wantErr:           false,
			wantAllTags:       []string{"1", "2", "4", "latest", "10"},
			wantNonHiddenTags: []string{"4", "latest", "10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Fake the client with the appropriate arguments
			client, fakeClientSet := occlient.FakeNew()
			fakeClientSet.ImageClientset.PrependReactor("list", "imagestreams", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, testingutil.FakeImageStreams(tt.args.name, tt.args.namespace, tt.args.tags), nil
			})
			fakeClientSet.ImageClientset.PrependReactor("list", "imagestreamtags", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, testingutil.FakeImageStreamTags(tt.args.name, tt.args.namespace, tt.args.tags, tt.args.hiddenTags), nil
			})

			// The function we are testing
			output, err := ListComponents(client)

			//Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component ListComponents() unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			// 1 call for current project + 1 call from openshift project for each of the ImageStream and ImageStreamTag
			if len(fakeClientSet.ImageClientset.Actions()) != 4 {
				t.Errorf("expected 2 ImageClientset.Actions() in ListComponents, got: %v", fakeClientSet.ImageClientset.Actions())
			}

			// Check if the output is the same as what's expected (for all tags)
			// and only if output is more than 0 (something is actually returned)
			if len(output.Items) > 0 && !(reflect.DeepEqual(output.Items[0].Spec.AllTags, tt.wantAllTags)) {
				t.Errorf("expected all tags: %s, got: %s", tt.wantAllTags, output.Items[0].Spec.AllTags)
			}

			// Check if the output is the same as what's expected (for hidden tags)
			// and only if output is more than 0 (something is actually returned)
			if len(output.Items) > 0 && !(reflect.DeepEqual(output.Items[0].Spec.NonHiddenTags, tt.wantNonHiddenTags)) {
				t.Errorf("expected non hidden tags: %s, got: %s", tt.wantNonHiddenTags, output.Items[0].Spec.NonHiddenTags)
			}

		})
	}
}

func TestSliceSupportedTags(t *testing.T) {

	imageStream := MockImageStream()

	img := ComponentType{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nodejs",
			Namespace: "openshift",
		},
		Spec: ComponentSpec{
			NonHiddenTags: []string{
				"12", "10", "8", "6", "latest",
			},
			ImageStreamTags: (*imageStream).Spec.Tags,
		},
	}

	supTags, unSupTags := SliceSupportedTags(img)
	if !reflect.DeepEqual(supTags, []string{"12", "10", "latest"}) ||
		!reflect.DeepEqual(unSupTags, []string{"8", "6"}) {
		t.Fatal("supported or unsupported tags are not as expected")
	}
}

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
  Experimental: true
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

func MockImageStream() *imagev1.ImageStream {

	tags := map[string]string{
		"12": "docker.io/rhscl/nodejs-12-rhel7:latest",
		"10": "docker.io/rhscl/nodejs-10-rhel7:latest",

		// unsupported ones
		"8": "docker.io/rhoar-nodejs/nodejs-8:latest",
		"6": "docker.io/rhoar-nodejs/nodejs-6:latest",
	}

	imageStream := &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nodejs",
			Namespace: "openshift",
		},
	}

	// this append is intentionally added before adding other tags
	// to confirm that tag references work even when they are out of order
	imageStream.Spec.Tags = append(imageStream.Spec.Tags,
		imagev1.TagReference{
			Name:        "latest",
			Annotations: map[string]string{"tags": "builder"},
			From: &corev1.ObjectReference{
				Kind: "ImageStreamTag",
				Name: "12",
			},
		})

	for tagName, imageName := range tags {
		imageTag := imagev1.TagReference{
			Name:        tagName,
			Annotations: map[string]string{"tags": "builder"},
			From: &corev1.ObjectReference{
				Kind: "DockerImage",
				Name: imageName,
			},
		}
		imageStream.Spec.Tags = append(imageStream.Spec.Tags, imageTag)
	}

	return imageStream
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
