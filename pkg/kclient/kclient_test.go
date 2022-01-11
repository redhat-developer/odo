package kclient

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/pkg/errors"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery/fake"
	ktesting "k8s.io/client-go/testing"
)

func (c *fakeDiscovery) ServerVersion() (*version.Info, error) {
	return &c.versionInfo, nil
}

type fakeDiscovery struct {
	*fake.FakeDiscovery
	versionInfo version.Info
}

func TestClient_IsSSASupported(t *testing.T) {

	tests := []struct {
		name    string
		version version.Info
		want    bool
	}{
		{
			name: "k8s with SSA",
			version: version.Info{
				Major:        "1",
				Minor:        "16",
				GitVersion:   "1.16.0+000000",
				GitCommit:    "",
				GitTreeState: "",
				BuildDate:    "",
				GoVersion:    runtime.Version(),
				Compiler:     runtime.Compiler,
				Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			},
			want: true,
		},
		{
			name: "k8s without SSA",
			version: version.Info{
				Major:        "1",
				Minor:        "15",
				GitVersion:   "1.15.0+000000",
				GitCommit:    "",
				GitTreeState: "",
				BuildDate:    "",
				GoVersion:    runtime.Version(),
				Compiler:     runtime.Compiler,
				Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			},
			want: false,
		},
		{
			name: "invalid k8s version",
			version: version.Info{
				Major:        "a",
				Minor:        "b",
				GitVersion:   "c",
				GitCommit:    "",
				GitTreeState: "",
				BuildDate:    "",
				GoVersion:    runtime.Version(),
				Compiler:     runtime.Compiler,
				Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			}, want: true,
		},
	}
	for _, tt := range tests {
		fkclient, _ := FakeNew()
		fd := fakeDiscovery{}
		fd.versionInfo = tt.version
		fkclient.SetDiscoveryInterface(&fd)

		t.Run(tt.name, func(t *testing.T) {
			if got := fkclient.IsSSASupported(); got != tt.want {
				t.Errorf("IsSSASupported() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	fkclient, fkclientset := FakeNew()
	fkclient.Namespace = "default"
	fkclientset.Kubernetes.PrependReactor("delete-collection", "deployments", func(action ktesting.Action) (bool, kruntime.Object, error) {
		if "a-selector=a-value" != action.(ktesting.DeleteCollectionAction).GetListRestrictions().Labels.String() {
			return true, nil, errors.New("not found")
		}
		return true, nil, nil
	})

	selectors := map[string]string{
		"a-selector": "a-value",
	}
	err := fkclient.Delete(selectors, false)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
}

// func TestKClient_CheckDefaultProject(t *testing.T) {
// 	tests := []struct {
// 		testName    string
// 		supported   bool
// 		wantErr     bool
// 		projectName string
// 		supportErr  error
// 	}{
// 		{
// 			testName:    "Case 0: CheckDefaultProject returns no error",
// 			projectName: "myproject",
// 			supported:   true,
// 			wantErr:     false,
// 			supportErr:  nil,
// 		},
// 		{
// 			testName:    "Case 1: CheckDefaultProject returns error on using 'default' project name in OC",
// 			projectName: "default",
// 			supported:   true,
// 			wantErr:     true,
// 			supportErr:  nil,
// 		},
// 		{testName: "Case 2: CheckDefaultProject returns error on checking if the project resource is supported",
// 			projectName: "myproject",
// 			supported:   false,
// 			wantErr:     true,
// 			supportErr:  fmt.Errorf("some error while checking project support"),
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.testName, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()
// 			mockKClient := NewMockClientInterface(ctrl)
// 			mockKClient.EXPECT().IsProjectSupported().Return(tt.supported, tt.supportErr).AnyTimes()
// 			//  TODO: This wil not work (pvala)
// 			got := mockKClient.CheckDefaultProject(tt.projectName)
// 			if (!tt.wantErr && got != nil) || (tt.wantErr && got == nil) {
// 				t.Errorf("got==nil: %v; wantErr: %v", got == nil, tt.wantErr)
// 			}
// 		})
// 	}
// }
