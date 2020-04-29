package config

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mkmik/multierror"
	"github.com/openshift/odo/pkg/manifest/ioutils"
	"knative.dev/pkg/apis"
)

const (
	DNS1035Error = "a DNS-1035 label must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character (e.g. 'my-name',  or 'abc-123', regex used for validation is '[a-z]([-a-z0-9]*[a-z0-9])?')"
)

func TestValidate(t *testing.T) {

	tests := []struct {
		desc string
		file string
		want error
	}{
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
	}

	for _, test := range tests {
		t.Run(test.desc, func(rt *testing.T) {
			manifest, err := ParseFile(ioutils.NewFilesystem(), test.file)
			if err != nil {
				rt.Fatalf("failed to parse file:%v", err)
			}
			got := manifest.Validate()
			err = matchMultiErrors(rt, got, test.want)
			if err != nil {
				rt.Fatal(err)
			}
		})
	}
}

func matchMultiErrors(t *testing.T, a error, b error) error {
	t.Helper()
	got, want := multierror.Split(a), multierror.Split(b)
	if len(got) != len(want) {
		return fmt.Errorf("did not match error, got %v want %v", got, want)
	}
	for i := 0; i < len(got); i++ {
		if diff := cmp.Diff(got[i].Error(), want[i].Error()); diff != "" {
			return fmt.Errorf("did not match error, got %v want %v", got[i], want[i])
		}
	}
	return nil
}
