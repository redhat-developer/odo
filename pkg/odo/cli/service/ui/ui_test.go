package ui

import (
	"github.com/Netflix/go-expect"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/redhat-developer/odo/pkg/testingutil"
	"github.com/stretchr/testify/require"
	"gopkg.in/AlecAivazis/survey.v1/core"
	"gopkg.in/AlecAivazis/survey.v1/terminal"
	"reflect"
	"testing"
)

func init() {
	// disable color output for all prompts to simplify testing
	core.DisableColor = true
}

func TestGetCategories(t *testing.T) {
	t.Run("getServiceClassesCategories should work", func(t *testing.T) {
		foo := testingutil.FakeClusterServiceClass("foo", "footag", "footag2")
		bar := testingutil.FakeClusterServiceClass("bar", "")
		boo := testingutil.FakeClusterServiceClass("boo")
		classes := map[string][]scv1beta1.ClusterServiceClass{"footag": {foo}, "other": {bar, boo}}

		categories := getServiceClassesCategories(classes)
		expected := []string{"footag", "other"}
		if !reflect.DeepEqual(expected, categories) {
			t.Errorf("test failed, expected %v, got %v", expected, categories)
		}
	})
}

func TestEnterServicePropertiesInteractively(t *testing.T) {
	// TODO: this test is currently skipped because it is not currently working properly. :(
	t.Skip("TODO: Skip this test until we can figure out what is wrong with it")

	tests := []struct {
		name           string
		servicePlan    scv1beta1.ClusterServicePlan
		expectedValues map[string]string
	}{
		{
			name:        "test 1 : with correct values",
			servicePlan: testingutil.FakeClusterServicePlan("dev", 1),
			expectedValues: map[string]string{
				"PLAN_DATABASE_URI":      "someuri",
				"PLAN_DATABASE_USERNAME": "default",
				"PLAN_DATABASE_PASSWORD": "foo",
			},
		},
	}

	for _, tt := range tests {
		plan := tt.servicePlan

		valuesPtr := new(map[string]string)
		testingutil.RunTest(t, func(c *expect.Console) {
			_, _ = c.ExpectString("Enter a value for string property PLAN_DATABASE_PASSWORD:")
			_, _ = c.SendLine("foo")
			_, _ = c.ExpectString("Enter a value for string property PLAN_DATABASE_URI:")
			_, _ = c.SendLine("")
			_, _ = c.ExpectString("Enter a value for string property PLAN_DATABASE_USERNAME:")
			_, _ = c.SendLine("")
			_, _ = c.ExpectString("Provide values for non-required properties")
			_, _ = c.SendLine("")
			_, _ = c.ExpectEOF()
		}, func(stdio terminal.Stdio) error {
			values := enterServicePropertiesInteractively(plan, stdio)
			valuesPtr = &values
			return nil
		})

		require.Equal(t, tt.expectedValues, *valuesPtr)
	}
}
