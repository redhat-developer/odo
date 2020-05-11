package ui

import (
	"reflect"
	"testing"

	"github.com/Netflix/go-expect"
	beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/openshift/odo/pkg/service"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/stretchr/testify/require"
	"gopkg.in/AlecAivazis/survey.v1/core"
	"gopkg.in/AlecAivazis/survey.v1/terminal"
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
		classes := map[string][]beta1.ClusterServiceClass{"footag": {foo}, "other": {bar, boo}}

		categories := getServiceClassesCategories(classes)
		expected := []string{"footag", "other"}
		if !reflect.DeepEqual(expected, categories) {
			t.Errorf("test failed, expected %v, got %v", expected, categories)
		}
	})
}

func TestGetServicePlanNames(t *testing.T) {
	t.Run("GetServicePlanNames should work", func(t *testing.T) {
		foo := testingutil.FakeClusterServicePlan("foo", 1)
		bar := testingutil.FakeClusterServicePlan("bar", 2)
		boo := testingutil.FakeClusterServicePlan("boo", 3)

		plans := GetServicePlanNames(map[string]beta1.ClusterServicePlan{"foo": foo, "bar": bar, "boo": boo})
		expected := []string{"bar", "boo", "foo"}
		if !reflect.DeepEqual(expected, plans) {
			t.Errorf("test failed, expected %v, got %v", expected, plans)
		}
	})
}

func TestWrapIfNeeded(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		prefixSize int
		expected   string
	}{
		{
			name:       "empty string, empty prefix",
			input:      "",
			prefixSize: 0,
			expected:   "",
		},
		{
			name:       "short string, empty prefix should not be wrapped",
			input:      "foo bar baz",
			prefixSize: 0,
			expected:   "foo bar baz",
		},
		{
			name:       "short string, empty prefix should not be wrapped with default width",
			input:      "foo bar baz",
			prefixSize: 2,
			expected:   "foo bar baz",
		},
		{
			name:       "short string, long prefix should wrap",
			input:      "foo bar baz",
			prefixSize: 78,
			expected:   "foo\nbar\nbaz",
		},
		{
			name:       "long string, empty prefix should wrap",
			input:      "0123456789 0123456789 0123456789 0123456789 0123456789 0123456789 0123456789 0123456789",
			prefixSize: 0,
			expected:   "0123456789 0123456789 0123456789 0123456789 0123456789 0123456789 0123456789\n0123456789",
		},
		{
			name:       "long string, short prefix should wrap",
			input:      "0123456789 0123456789 0123456789 0123456789 0123456789 0123456789 0123456789 0123456789",
			prefixSize: 2,
			expected:   "0123456789 0123456789 0123456789 0123456789 0123456789 0123456789 0123456789\n0123456789",
		},
		{
			name:       "long string, longer prefix should wrap more",
			input:      "0123456789 0123456789 0123456789 0123456789 0123456789 0123456789 0123456789 0123456789",
			prefixSize: 10,
			expected:   "0123456789 0123456789 0123456789 0123456789 0123456789 0123456789\n0123456789 0123456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := wrapIfNeeded(tt.input, tt.prefixSize)
			if tt.expected != output {
				t.Errorf("test failed, expected %s, got %s", tt.expected, output)
			}
		})
	}
}

func init() {
	// disable color output for all prompts to simplify testing
	core.DisableColor = true
}

func TestEnterServicePropertiesInteractively(t *testing.T) {
	// TODO: this test is currently skipped because it is not currently working properly. :(
	t.Skip("TODO: Skip this test until we can figure out what is wrong with it")

	tests := []struct {
		name           string
		servicePlan    beta1.ClusterServicePlan
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

func TestGetLongDescription(t *testing.T) {
	desc := testingutil.FakeClusterServiceClass("foo")
	desc.Spec.ExternalMetadata = testingutil.SingleValuedRawExtension("longDescription", "description")
	empty := testingutil.FakeClusterServiceClass("foo")
	empty.Spec.ExternalMetadata = testingutil.SingleValuedRawExtension("longDescription", "")
	tests := []struct {
		name     string
		input    beta1.ClusterServiceClass
		expected string
	}{
		{
			name:     "no metadata",
			input:    testingutil.FakeClusterServiceClass("foo"),
			expected: "",
		},
		{
			name:     "description",
			input:    desc,
			expected: "description",
		},
		{
			name:     "empty description",
			input:    empty,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := getLongDescription(tt.input)
			if tt.expected != output {
				t.Errorf("test failed, expected %s, got %s", tt.expected, output)
			}
		})
	}
}

func TestPropDesc(t *testing.T) {
	tests := []struct {
		name     string
		prop     service.ServicePlanParameter
		expected string
	}{
		{
			name:     "empty",
			prop:     service.ServicePlanParameter{},
			expected: "",
		},
		{
			name:     "name only",
			prop:     service.ServicePlanParameter{Name: "foo"},
			expected: "foo",
		},
		{
			name:     "with title",
			prop:     service.ServicePlanParameter{Name: "foo", Title: "title"},
			expected: "foo (title)",
		},
		{
			name:     "with description",
			prop:     service.ServicePlanParameter{Name: "foo", Description: "desc"},
			expected: "foo (desc)",
		},
		{
			name:     "with title and description",
			prop:     service.ServicePlanParameter{Name: "foo", Description: "desc", Title: "title"},
			expected: "foo (title)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := propDesc(tt.prop)
			if tt.expected != output {
				t.Errorf("test failed, expected %v, got %v", tt.expected, output)
			}
		})
	}
}
