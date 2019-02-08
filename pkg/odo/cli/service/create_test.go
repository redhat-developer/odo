package service

import (
	"testing"
)

func TestOutputNonInteractiveEquivalent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		options  ServiceCreateOptions
		expected string
	}{
		{
			name: "when output is not requested, should return empty string",
			options: ServiceCreateOptions{
				CmdFullName: RecommendedCommandName,
				outputCLI:   false,
				ServiceType: "foo",
			},
			expected: "",
		},
		{
			name: "just service class",
			options: ServiceCreateOptions{
				CmdFullName: RecommendedCommandName,
				outputCLI:   true,
				ServiceType: "foo",
			},
			expected: RecommendedCommandName + " foo",
		},
		{
			name: "just service class and name",
			options: ServiceCreateOptions{
				CmdFullName: RecommendedCommandName,
				outputCLI:   true,
				ServiceType: "foo",
				ServiceName: "myservice",
			},
			expected: RecommendedCommandName + " foo myservice",
		},
		{
			name: "service class, name and plan",
			options: ServiceCreateOptions{
				CmdFullName: RecommendedCommandName,
				outputCLI:   true,
				ServiceType: "foo",
				ServiceName: "myservice",
				Plan:        "dev",
			},
			expected: RecommendedCommandName + " foo myservice --plan dev",
		},
		{
			name: "service class and plan",
			options: ServiceCreateOptions{
				CmdFullName: RecommendedCommandName,
				outputCLI:   true,
				ServiceType: "foo",
				Plan:        "dev",
			},
			expected: RecommendedCommandName + " foo --plan dev",
		},
		{
			name: "service class and empty params",
			options: ServiceCreateOptions{
				CmdFullName:   RecommendedCommandName,
				outputCLI:     true,
				ServiceType:   "foo",
				ParametersMap: map[string]string{},
			},
			expected: RecommendedCommandName + " foo",
		},
		{
			name: "service class and params",
			options: ServiceCreateOptions{
				CmdFullName:   RecommendedCommandName,
				outputCLI:     true,
				ServiceType:   "foo",
				ParametersMap: map[string]string{"param1": "value1", "param2": "value2"},
			},
			expected: RecommendedCommandName + " foo -p param1=value1 -p param2=value2",
		},
		{
			name: "all",
			options: ServiceCreateOptions{
				CmdFullName:   RecommendedCommandName,
				outputCLI:     true,
				ServiceType:   "foo",
				ServiceName:   "name",
				Plan:          "plan",
				ParametersMap: map[string]string{"param1": "value1", "param2": "value2"},
			},
			expected: RecommendedCommandName + " foo name --plan plan -p param1=value1 -p param2=value2",
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
