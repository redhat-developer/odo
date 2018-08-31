package catalog

import (
	"reflect"
	"testing"

	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/testingutil"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestVersionExist(t *testing.T) {
	type args struct {
		name             string
		namespace        string
		componentType    string
		componentVersion string
		tags             []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Case 1: The version exists (alphabetical)",
			args: args{
				name:             "nodejs",
				namespace:        "openshift",
				componentType:    "nodejs",
				componentVersion: "dev",
				tags:             []string{"latest", "1.0.0", "test", "dev"},
			},
			wantErr: false,
		},
		{
			name: "Case 2: The version exists (number)",
			args: args{
				name:             "nodejs",
				namespace:        "openshift",
				componentType:    "nodejs",
				componentVersion: "1.0.0",
				tags:             []string{"0.0.1", "1.0.0", "9999", "0.0.1"},
			},
			wantErr: false,
		},
		{
			name: "Case 3: The version does not exist",
			args: args{
				name:             "nodejs",
				namespace:        "openshift",
				componentType:    "nodejs",
				componentVersion: "latest",
				tags:             []string{"foobar"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Fake the client with the appropriate arguments
			client, fakeClientSet := occlient.FakeNew()
			fakeClientSet.ImageClientset.PrependReactor("list", "imagestreams", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, testingutil.FakeImageStreams(tt.args.name, tt.args.namespace, tt.args.tags), nil
			})

			// The function we are testing
			doesItExist, err := VersionExists(client, tt.args.componentType, tt.args.componentVersion)

			if err != nil {
				t.Errorf("VersionExist() errored when it shouldn't have: %s", err)
			}

			// Checks for error in positive cases
			if tt.wantErr && doesItExist {
				t.Errorf("VersionExist() unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if len(fakeClientSet.ImageClientset.Actions()) != 2 { // 1 call for current project + 1 call from openshift project
				t.Errorf("expected 2 ImageClientset.Actions() in VersionExist, got: %v", fakeClientSet.ImageClientset.Actions())
			}

			// Check if the output is the same as what's expected (tags)
			// and only if output is more than 0 (something is actually returned)
			if !tt.wantErr && !doesItExist {
				t.Errorf("VersionExist() unexpected tag. Expected tag %s", tt.args.componentVersion)
			}

		})
	}

}

func TestList(t *testing.T) {
	type args struct {
		name      string
		namespace string
		tags      []string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantTags []string
	}{
		{
			name: "Case 1: Valid image output with one tag",
			args: args{
				name:      "foobar",
				namespace: "openshift",
				tags:      []string{"latest"},
			},
			wantErr:  false,
			wantTags: []string{"latest"},
		},
		{
			name: "Case 2: Valid image output with multiple tags",
			args: args{
				name:      "foobar",
				namespace: "openshift",
				tags:      []string{"1.0.0", "1.0.1", "0.0.1", "latest"},
			},
			wantErr:  false,
			wantTags: []string{"1.0.0", "1.0.1", "0.0.1", "latest"},
		},
		{
			name: "Case 3: Invalid image output with no tags",
			args: args{
				name:      "foobar",
				namespace: "foo",
				tags:      []string{},
			},
			wantErr:  true,
			wantTags: []string{},
		},
		{
			name: "Case 4: Valid image with output tags from a different namespace",
			args: args{
				name:      "foobar",
				namespace: "foo",
				tags:      []string{"1.0.0", "1.0.1", "0.0.1", "latest"},
			},
			wantErr:  false,
			wantTags: []string{"1.0.0", "1.0.1", "0.0.1", "latest"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Fake the client with the appropriate arguments
			client, fakeClientSet := occlient.FakeNew()
			fakeClientSet.ImageClientset.PrependReactor("list", "imagestreams", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, testingutil.FakeImageStreams(tt.args.name, tt.args.namespace, tt.args.tags), nil
			})

			// The function we are testing
			output, err := List(client)

			//Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component List() unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if len(fakeClientSet.ImageClientset.Actions()) != 2 { // 1 call for current project + 1 call from openshift project
				t.Errorf("expected 2 ImageClientset.Actions() in List, got: %v", fakeClientSet.ImageClientset.Actions())
			}

			// Check if the output is the same as what's expected (tags)
			// and only if output is more than 0 (something is actually returned)
			if len(output) > 0 && !(reflect.DeepEqual(output[0].Tags, tt.wantTags)) {
				t.Errorf("expected tags: %s, got: %s", tt.wantTags, output[0].Tags)
			}

		})
	}
}
