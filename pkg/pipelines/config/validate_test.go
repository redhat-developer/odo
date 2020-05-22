package config

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mkmik/multierror"
	"github.com/openshift/odo/pkg/pipelines/ioutils"
	"knative.dev/pkg/apis"
)

const (
	DNS1035Error = "a DNS-1035 label must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character (e.g. 'my-name',  or 'abc-123', regex used for validation is '[a-z]([-a-z0-9]*[a-z0-9])?')"
)

func TestValidate(t *testing.T) {

	tests := []struct {
		desc     string
		filename string
		wantErr  error
	}{
		{
			"cicd environment cannot contain applications and services",
			"testdata/cicd_env_cant_have_apps_svcs.yaml",
			multierror.Join(
				[]error{
					invalidEnvironment("test-cicd", "A special environment cannot contain services.", []string{"environments.test-cicd.services.bus-svc"}),
					invalidEnvironment("test-cicd", "A special environment cannot contain applications.", []string{"environments.test-cicd.apps.bus"}),
				},
			),
		},
		{
			"argocd environment cannot contain applications and services",
			"testdata/argocd_env_cant_have_apps_svcs.yaml",
			multierror.Join(
				[]error{
					invalidEnvironment("test-argocd", "A special environment cannot contain services.", []string{"environments.test-argocd.services.bus-svc"}),
					invalidEnvironment("test-argocd", "A special environment cannot contain applications.", []string{"environments.test-argocd.apps.bus"}),
				},
			),
		},
		{
			"Invalid entity name error",
			"testdata/name_error.yaml",
			multierror.Join(
				[]error{
					invalidNameError("", DNS1035Error, []string{"environments.develo.pment.services"}),
					invalidNameError("", DNS1035Error, []string{"environments.develo.pment.services.pipelines.integration.binding"}),
					invalidNameError("app-1$.", DNS1035Error, []string{"environments.develo.pment.apps.app-1$."}),
					invalidNameError("develo.pment", DNS1035Error, []string{"environments.develo.pment"}),
					invalidNameError("test)cicd", DNS1035Error, []string{"environments.test)cicd"}),
				},
			),
		},
		{
			"Missing field error",
			"testdata/missing_fields_error.yaml",
			multierror.Join([]error{
				missingFieldsError([]string{"secret"}, []string{"environments.development.services.service-1.webhook"}),
				missingFieldsError([]string{"integration"}, []string{"environments.development.services.service-1.pipelines"}),
			}),
		},
		{
			"Missing service and config repo from application",
			"testdata/missing_service_error.yaml",
			multierror.Join([]error{
				missingFieldsError([]string{"services", "config_repo"}, []string{"environments.development.apps.app-1"}), missingFieldsError([]string{"path"}, []string{"environments.development.apps.app-2.config_repo"}),
				missingFieldsError([]string{"url"}, []string{"environments.development.apps.app-3.config_repo"}),
				missingFieldsError([]string{"url", "path"}, []string{"environments.development.apps.app-4.config_repo"}),
				apis.ErrMultipleOneOf("environments.development.apps.app-5.services", "environments.development.apps.app-5.config_repo"),
			}),
		},
		{
			"duplicate environment name error",
			"testdata/duplicate_environment.yaml",
			multierror.Join(
				[]error{
					duplicateFieldsError([]string{"duplicate-environment"}, []string{"environments.duplicate-environment"}),
				},
			),
		},
		{
			"duplicate application name error",
			"testdata/duplicate_application.yaml",
			multierror.Join(
				[]error{
					duplicateFieldsError([]string{"my-app-1"}, []string{"environments.app-environment.apps.my-app-1"}),
				},
			),
		},
		{
			"duplicate service name error",
			"testdata/duplicate_service.yaml",
			multierror.Join(
				[]error{
					duplicateFieldsError([]string{"app-1-service-http"}, []string{"environments.duplicate-service.services.app-1-service-http"}),
				},
			),
		},
		{
			"missing app service reference",
			"testdata/missing_service_in_application.yaml",
			multierror.Join(
				[]error{
					missingServiceRefError("app-1-svc-http", "my-app-1", []string{"environments.duplicate-service.apps.my-app-1"}),
				},
			),
		},
		{
			"missing app service reference",
			"testdata/duplicate_source_url.yaml",
			multierror.Join(
				[]error{
					duplicateSourceError("https://github.com/testing/testing.git", []string{"environments.duplicate-source.services.app-1-service-http", "environments.duplicate-source.services.app-2-service-http"}),
				},
			),
		},
		{
			"service with pipeline with no template",
			"testdata/service_with_bindings_no_template.yaml",
			nil,
		},
		{
			"valid manifest file",
			"testdata/valid_manifest.yaml",
			nil,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s (%s)", test.desc, test.filename), func(rt *testing.T) {
			pipelines, err := ParseFile(ioutils.NewFilesystem(), test.filename)
			if err != nil {
				rt.Fatalf("failed to parse file:%v", err)
			}
			got := pipelines.Validate()
			err = matchMultiErrors(rt, got, test.wantErr)
			if err != nil {
				rt.Fatal(err)
			}
		})
	}
}

func matchMultiErrors(t *testing.T, a error, b error) error {
	t.Helper()
	if a == nil || b == nil {
		if a != b {
			return fmt.Errorf("did not match error, got %v want %v", a, b)
		}
		return nil
	}
	got, want := multierror.Split(a), multierror.Split(b)
	if len(got) != len(want) {
		return fmt.Errorf("error count did not match, got %d want %d", len(got), len(want))
	}
	for i := 0; i < len(got); i++ {
		if diff := cmp.Diff(got[i].Error(), want[i].Error()); diff != "" {
			return fmt.Errorf("did not match error, got %v want %v", got[i], want[i])
		}
	}
	return nil
}
