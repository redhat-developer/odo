package kclient

import (
	"fmt"
	"runtime"
	"testing"

	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery/fake"
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
