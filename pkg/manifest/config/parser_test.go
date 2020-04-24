package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	parseTests := []struct {
		filename string
		want     *Manifest
	}{
		{"testdata/example1.yaml", &Manifest{
			Environments: []*Environment{
				{
					Name: "development",
					Pipelines: &Pipelines{
						Integration: &TemplateBinding{
							Template: "dev-ci-template",
							Binding:  "dev-ci-binding",
						},
					},
					Apps: []*Application{
						{
							Name: "my-app-1",
							Services: []*Service{
								{
									Name:      "app-1-service-http",
									SourceURL: "https://github.com/myproject/myservice.git",
								},
								{Name: "app-1-service-metrics"},
							},
						},
						{
							Name: "my-app-2",
							Services: []*Service{
								{Name: "app-2-service"},
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
					Apps: []*Application{
						{Name: "my-app-1",
							Services: []*Service{
								{Name: "app-1-service-http"},
								{Name: "app-1-service-metrics"},
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
					Apps: []*Application{
						{
							Name: "my-app-1",
							Services: []*Service{
								{
									Name:      "app-1-service-http",
									SourceURL: "https://github.com/myproject/myservice.git",
								},
								{Name: "app-1-service-metrics"},
							},
						},
					},
				},
				{
					Name:   "tst-cicd",
					IsCICD: true,
				},
			},
		},
		},
	}

	for _, tt := range parseTests {
		t.Run(fmt.Sprintf("parsing %s", tt.filename), func(rt *testing.T) {
			f, err := os.Open(tt.filename)
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
