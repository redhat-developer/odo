package catalog

import (
	"reflect"
	"testing"

	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/testingutil"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

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

			// Check if the output is the same as what's expected (tags)
			// and only if output is more than 0 (something is actually returned)
			if len(output) > 0 && !(reflect.DeepEqual(output[0].Tags, tt.wantTags)) {
				t.Errorf("expected tags: %s, got: %s", tt.wantTags, output[0].Tags)
			}

		})
	}
}
