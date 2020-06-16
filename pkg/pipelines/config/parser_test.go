package config

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openshift/odo/pkg/pipelines/ioutils"
)

func TestParse(t *testing.T) {
	parseTests := []struct {
		filename string
		want     *Manifest
	}{
		{"testdata/example1.yaml", &Manifest{
			Config: &Config{
				Pipelines: &PipelinesConfig{
					Name: "test-pipelines",
				},
				ArgoCD: &ArgoCDConfig{
					Namespace: "test-argocd",
				},
			},
			Environments: []*Environment{
				{
					Name: "development",
					Pipelines: &Pipelines{
						Integration: &TemplateBinding{
							Template: "dev-ci-template",
							Bindings: []string{"dev-ci-binding"},
						},
					},
					Services: []*Service{
						{
							Name:      "service-http",
							SourceURL: "https://github.com/myproject/myservice.git",
						},
						{Name: "service-redis"},
					},
					Apps: []*Application{
						{
							Name: "my-app-1",
							ServiceRefs: []string{
								"service-http",
							},
						},
						{
							Name: "my-app-2",
							ServiceRefs: []string{
								"service-redis",
							},
						},
					},
				},
				{
					Name: "staging",
					Apps: []*Application{
						{Name: "my-app-1",
							ConfigRepo: &Repository{
								URL:            "https://github.com/testing/testing",
								TargetRevision: "master",
								Path:           "config",
							},
						},
					},
				},
				{
					Name: "production",
					Services: []*Service{
						{Name: "service-http"},
						{Name: "service-metrics"},
					},
					Apps: []*Application{
						{
							Name: "my-app-1",
							ServiceRefs: []string{
								"service-http",
								"service-metrics",
							},
						},
					},
				},
			},
		},
		},

		{"testdata/example2.yaml", &Manifest{
			Environments: []*Environment{
				{
					Name: "development",
					Services: []*Service{
						{
							Name:      "app-1-service-http",
							SourceURL: "https://github.com/myproject/myservice.git",
						},
						{Name: "app-1-service-metrics"},
					},
					Apps: []*Application{
						{
							Name: "my-app-1",
							ServiceRefs: []string{
								"app-1-service-http",
								"app-1-service-metrics",
							},
						},
					},
				},
				{
					Name: "tst-cicd",
				},
			},
		},
		},
		{"testdata/example-with-cluster.yaml", &Manifest{
			Environments: []*Environment{
				{
					Name:    "development",
					Cluster: "testing.cluster",
					Services: []*Service{
						{Name: "service-http", SourceURL: "https://github.com/myproject/myservice.git"},
					},
					Apps: []*Application{
						{Name: "my-app-1", ServiceRefs: []string{"service-http"}},
					},
				},
			},
		},
		},
	}

	for _, tt := range parseTests {
		t.Run(fmt.Sprintf("parsing %s", tt.filename), func(rt *testing.T) {
			fs := ioutils.NewFilesystem()
			f, err := fs.Open(tt.filename)
			if err != nil {
				rt.Fatalf("failed to open %v: %s", tt.filename, err)
			}
			defer f.Close()

			got, err := Parse(f)
			if err != nil {
				rt.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				rt.Errorf("Parse(%s) failed diff\n%s", tt.filename, diff)
			}
		})
	}
}
