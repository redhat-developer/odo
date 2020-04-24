package config

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mkmik/multierror"
)

const (
	DNS1035Error = "[a DNS-1035 label must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character (e.g. 'my-name',  or 'abc-123', regex used for validation is '[a-z]([-a-z0-9]*[a-z0-9])?')]"
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
			multierror.Join([]error{
				fmt.Errorf("%q is not a valid name at %v: \n%v", "", "environments/develo.pment/services", DNS1035Error),
				fmt.Errorf("%q is not a valid name at %v: \n%v", "", "environments/develo.pment/services/pipelines", DNS1035Error),
				fmt.Errorf("%q is not a valid name at %v: \n%v", "app-1$.", "environments/develo.pment/apps/app-1$.", DNS1035Error),
				fmt.Errorf("%q is not a valid name at %v: \n%v", "develo.pment", "environments/develo.pment", DNS1035Error),
				fmt.Errorf("%q is not a valid name at %v: \n%v", "test)cicd", "environments/test)cicd", DNS1035Error),
			}),
		},
		{
			"Missing field error",
			"testdata/missing_fields_error.yaml",
			multierror.Join([]error{
				fmt.Errorf("%v not found at %v", "secret", "environments/development/services/service-1"),
				fmt.Errorf("%v not found at %v", "pipelines", "environments/development/services/service-1"),
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(rt *testing.T) {
			manifest, err := ParseFile(test.file)
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
