package webhook

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openshift/odo/pkg/pipelines/config"
)

func TestBuildURL(t *testing.T) {

	testcases := []struct {
		host   string
		hasTLS bool
		want   string
	}{
		{
			host:   "test.example.com",
			hasTLS: false,
			want:   "http://test.example.com",
		},
		{
			host:   "test.example.com",
			hasTLS: true,
			want:   "https://test.example.com",
		},
	}

	for i, tt := range testcases {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			got := buildURL(tt.host, tt.hasTLS)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("result mismatch got\n%s", diff)
			}
		})
	}
}
func TestGetGitRepoURL(t *testing.T) {

	testcases := []struct {
		manifest    *config.Manifest
		isCICD      bool
		serviceName *QualifiedServiceName
		want        string
	}{
		{
			manifest: &config.Manifest{
				GitOpsURL: "https://github.com/foo/bar.git",
			},
			isCICD: true,
			want:   "https://github.com/foo/bar.git",
		},
		{
			manifest: &config.Manifest{},
			want:     "",
		},
		{
			manifest: &config.Manifest{
				GitOpsURL: "https://github.com/foo/bar.git",
				Environments: []*config.Environment{
					{
						Name: "notmyenv",
						Apps: []*config.Application{
							{
								Name: "notmyapp",
								Services: []*config.Service{
									{
										Name:      "notmyservice",
										SourceURL: "https://not/mine",
									},
								},
							},
						},
					},
					{
						Name: "myenv",
						Apps: []*config.Application{
							{
								Name: "notmyapp",
								Services: []*config.Service{
									{
										Name:      "notmyservice",
										SourceURL: "https://not/mine",
									},
								},
							},
							{
								Name: "myapp",
								Services: []*config.Service{
									{
										Name:      "notmyservice",
										SourceURL: "https://not/mine",
									},
									{
										Name:      "myservice",
										SourceURL: "https://github.com/foo2/bar.git",
									},
									{
										Name:      "notmyserviceagain",
										SourceURL: "https://not/mine",
									},
								},
							},
							{
								Name: "notmyapp2",
								Services: []*config.Service{
									{
										Name:      "notmyservice",
										SourceURL: "https://not/mine",
									},
								},
							},
						},
					},
				},
			},
			isCICD:      false,
			serviceName: &QualifiedServiceName{EnvironmentName: "myenv", ApplicationName: "myapp", ServiceName: "myservice"},
			want:        "https://github.com/foo2/bar.git",
		},
	}

	for i, tt := range testcases {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			got := getRepoURL(tt.manifest, tt.isCICD, tt.serviceName)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("result mismatch got\n%s", diff)
			}
		})
	}
}
