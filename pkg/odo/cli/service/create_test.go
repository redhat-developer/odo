package service

import (
	"testing"

	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
)

func TestOutputNonInteractiveEquivalent(t *testing.T) {
	t.Parallel()

	client, _ := occlient.FakeNew()

	tests := []struct {
		name     string
		options  ServiceCreateOptions
		expected string
	}{
		{
			name: "when output is not requested, should return empty string",
			options: ServiceCreateOptions{
				Context: genericclioptions.NewFakeContext("testproject", "app", "", client, nil),

				CmdFullName: RecommendedCommandName,
				outputCLI:   false,
				ServiceType: "foo",
			},
			expected: "",
		},
		{
			name: "just service class",
			options: ServiceCreateOptions{
				Context: genericclioptions.NewFakeContext("testproject", "app", "", client, nil),

				CmdFullName: RecommendedCommandName,
				outputCLI:   true,
				ServiceType: "foo",
			},
			expected: RecommendedCommandName + " foo --app app --project testproject",
		},
		{
			name: "just service class and name",
			options: ServiceCreateOptions{
				Context: genericclioptions.NewFakeContext("testproject", "app", "", client, nil),

				CmdFullName: RecommendedCommandName,
				outputCLI:   true,
				ServiceType: "foo",
				ServiceName: "myservice",
			},
			expected: RecommendedCommandName + " foo myservice --app app --project testproject",
		},
		{
			name: "service class, name and plan",
			options: ServiceCreateOptions{
				Context: genericclioptions.NewFakeContext("testproject", "app", "", client, nil),

				CmdFullName: RecommendedCommandName,
				outputCLI:   true,
				ServiceType: "foo",
				ServiceName: "myservice",
				Plan:        "dev",
			},
			expected: RecommendedCommandName + " foo myservice --app app --project testproject --plan dev",
		},
		{
			name: "service class and plan",
			options: ServiceCreateOptions{
				Context:     genericclioptions.NewFakeContext("testproject", "app", "", client, nil),
				CmdFullName: RecommendedCommandName,
				outputCLI:   true,
				ServiceType: "foo",
				Plan:        "dev",
			},
			expected: RecommendedCommandName + " foo --app app --project testproject --plan dev",
		},
		{
			name: "service class and empty params",
			options: ServiceCreateOptions{
				Context:       genericclioptions.NewFakeContext("testproject", "app", "", client, nil),
				CmdFullName:   RecommendedCommandName,
				outputCLI:     true,
				ServiceType:   "foo",
				ParametersMap: map[string]string{},
			},
			expected: RecommendedCommandName + " foo --app app --project testproject",
		},
		{
			name: "service class and params",
			options: ServiceCreateOptions{
				Context:       genericclioptions.NewFakeContext("testproject", "app", "", client, nil),
				CmdFullName:   RecommendedCommandName,
				outputCLI:     true,
				ServiceType:   "foo",
				ParametersMap: map[string]string{"param1": "value1", "param2": "value2"},
			},
			expected: RecommendedCommandName + " foo --app app --project testproject -p param1=value1 -p param2=value2",
		},
		{
			name: "all",
			options: ServiceCreateOptions{
				Context:       genericclioptions.NewFakeContext("testproject", "app", "", client, nil),
				CmdFullName:   RecommendedCommandName,
				outputCLI:     true,
				ServiceType:   "foo",
				ServiceName:   "name",
				Plan:          "plan",
				ParametersMap: map[string]string{"param1": "value1", "param2": "value2"},
			},
			expected: RecommendedCommandName + " foo name --app app --project testproject --plan plan -p param1=value1 -p param2=value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.options.outputNonInteractiveEquivalent()
			if tt.expected != actual {
				t.Errorf("expected '%s', got '%s'", tt.expected, actual)
			}
		})
	}

}
