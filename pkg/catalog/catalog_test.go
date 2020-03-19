package catalog

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	imagev1 "github.com/openshift/api/image/v1"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/testingutil"
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
				"12", "10", "8", "latest",
			},
			ImageStreamRef: *imageStream,
		},
	}

	supTags, unSupTags := SliceSupportedTags(img)

	if !reflect.DeepEqual(supTags, []string{"12", "10", "latest"}) ||
		!reflect.DeepEqual(unSupTags, []string{"8"}) {
		t.Fatal("supported or unsupported tags are not as expected")
	}
}

func TestGetDevfileIndex(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		_, err := rw.Write([]byte(
			`
			[
				{
					"displayName": "NodeJS Angular Web Application",
					"description": "Stack for developing NodeJS Angular Web Application",
					"tags": [
						"NodeJS",
						"Angular",
						"Alpine"
					],
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

	tests := []struct {
		name             string
		devfileIndexLink string
		want             []DevfileIndexEntry
	}{
		{
			name:             "Test NodeJS devfile index",
			devfileIndexLink: server.URL,
			want: []DevfileIndexEntry{
				{
					DisplayName: "NodeJS Angular Web Application",
					Description: "Stack for developing NodeJS Angular Web Application",
					Tags: []string{
						"NodeJS",
						"Angular",
						"Alpine",
					},
					Icon:              "/images/angular.svg",
					GlobalMemoryLimit: "2686Mi",
					Links: struct {
						Link string `json:"self"`
					}{
						Link: "/devfiles/angular/devfile.yaml",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDevfileIndex(tt.devfileIndexLink)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got: %v, want: %v", got, tt.want)
				t.Logf("Error message is: %v", err)
			}
		})
	}
}

func TestGetDevfile(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		// Note: Yaml file uses indentation to represent relationships between data layers,
		// so we need to use the following Yaml format to obey the rule
		_, err := rw.Write([]byte(
			`apiVersion: 1.0.0
metadata:
  generateName: angular-
  projects:
  -
    name: angular-realworld-example-app
    source:
      type: git
      location: "https://github.com/gothinkster/angular-realworld-example-app"
components:
  -
    type: chePlugin
    id: che-incubator/typescript/latest
  -
    type: dockerimage
    alias: nodejs
    image: quay.io/eclipse/che-nodejs10-community:nightly
    memoryLimit: 1Gi
    endpoints:
      - name: 'angular'
        port: 4200
    mountSources: true
commands:
  - name: yarn install
    actions:
      - type: exec
        component: nodejs
        command: yarn install
        workdir: ${CHE_PROJECTS_ROOT}/angular-realworld-example-app
  -
    name: build
    actions:
      - type: exec
        component: nodejs
        command: yarn run build
        workdir: ${CHE_PROJECTS_ROOT}/angular-realworld-example-app
  -
    name: start
    actions:
      - type: exec
        component: nodejs
        command: yarn run start --host 0.0.0.0 --disableHostCheck true
        workdir: ${CHE_PROJECTS_ROOT}/angular-realworld-example-app
  -
    name: lint
    actions:
      - type: exec
        component: nodejs
        command: yarn run lint
        workdir: ${CHE_PROJECTS_ROOT}/angular-realworld-example-app`,
		))
		if err != nil {
			t.Error(err)
		}
	}))
	// Close the server when test finishes
	defer server.Close()

	tests := []struct {
		name        string
		devfileLink string
		want        Devfile
	}{
		{
			name:        "Test NodeJS devfile",
			devfileLink: server.URL,
			want: Devfile{
				APIVersion: "1.0.0",
				MetaData: struct {
					GenerateName string `yaml:"generateName"`
				}{
					GenerateName: "angular-",
				},
				Components: []struct {
					Type  string `yaml:"type"`
					Alias string `yaml:"alias"`
				}{
					{
						Type: "chePlugin",
					},
					{
						Type:  "dockerimage",
						Alias: "nodejs",
					},
				},
				Commands: []struct {
					Name string `yaml:"name"`
				}{
					{
						Name: "yarn install",
					},
					{
						Name: "build",
					},
					{
						Name: "start",
					},
					{
						Name: "lint",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDevfile(tt.devfileLink)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got: %v, want: %v", got, tt.want)
				t.Logf("Error message is: %v", err)
			}
		})
	}
}

func TestIsDevfileComponentSupported(t *testing.T) {
	tests := []struct {
		name    string
		devfile Devfile
		want    bool
	}{
		{
			name: "Case 1: Test unsupported devfile",
			devfile: Devfile{
				APIVersion: "1.0.0",
				MetaData: struct {
					GenerateName string `yaml:"generateName"`
				}{
					GenerateName: "angular-",
				},
				Components: []struct {
					Type  string `yaml:"type"`
					Alias string `yaml:"alias"`
				}{
					{
						Type: "chePlugin",
					},
					{
						Type:  "dockerimage",
						Alias: "nodejs",
					},
				},
				Commands: []struct {
					Name string `yaml:"name"`
				}{
					{
						Name: "yarn install",
					},
					{
						Name: "build",
					},
					{
						Name: "start",
					},
					{
						Name: "lint",
					},
				},
			},
			want: false,
		},
		{
			name: "Case 2: Test supported devfile",
			devfile: Devfile{
				APIVersion: "1.0.0",
				MetaData: struct {
					GenerateName string `yaml:"generateName"`
				}{
					GenerateName: "openLiberty",
				},
				Components: []struct {
					Type  string `yaml:"type"`
					Alias string `yaml:"alias"`
				}{
					{
						Type: "chePlugin",
					},
					{
						Type:  "dockerimage",
						Alias: "runtime",
					},
				},
				Commands: []struct {
					Name string `yaml:"name"`
				}{
					{
						Name: "devBuild",
					},
					{
						Name: "devRun",
					},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDevfileComponentSupported(tt.devfile)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got: %v, want: %v", got, tt.want)
			}
		})
	}
}

func MockImageStream() *imagev1.ImageStream {

	tags := map[string]string{
		"12": "docker.io/rhscl/nodejs-12-rhel7:latest",
		"10": "docker.io/rhscl/nodejs-10-rhel7:latest",
		"8":  "docker.io/rhoar-nodejs/nodejs-8:latest",
	}

	imageStream := &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nodejs",
			Namespace: "openshift",
		},
	}

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

	imageStream.Spec.Tags = append(imageStream.Spec.Tags,
		imagev1.TagReference{
			Name:        "latest",
			Annotations: map[string]string{"tags": "builder"},
			From: &corev1.ObjectReference{
				Kind: "ImageStreamTag",
				Name: "12",
			},
		})

	return imageStream
}
