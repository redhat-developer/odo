package env

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestPrintSupportedParameters(t *testing.T) {
	supportedSetParameters := map[string]string{
		nameParameter:      nameParameterDescription,
		namespaceParameter: namespaceParameterDescription,
		debugportParameter: debugportParameterDescription,
	}

	wantSetParameters := `Available parameters:
  DebugPort: Use this value to set component debug port
  Name: Use this value to set component name
  Namespace: Use this value to set component namespace`

	supportedUnsetParameters := map[string]string{
		debugportParameter: debugportParameterDescription,
	}

	wantUnsetParameters := `Available parameters:
  DebugPort: Use this value to set component debug port`

	tests := []struct {
		name                string
		supportedParameters map[string]string
		want                string
	}{
		{
			name:                "Case 1: Test print supported set parameters",
			supportedParameters: supportedSetParameters,
			want:                wantSetParameters,
		},
		{
			name:                "Case 2: Test print supported unset parameters",
			supportedParameters: supportedUnsetParameters,
			want:                wantUnsetParameters,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strings.TrimSpace(printSupportedParameters(tt.supportedParameters))

			gotStrings := strings.Split(got, "\n")
			wantStrings := strings.Split(tt.want, "\n")

			sort.Strings(gotStrings)
			sort.Strings(wantStrings)

			if !reflect.DeepEqual(wantStrings, gotStrings) {
				t.Errorf("\nGot: %s\nWant: %s", got, tt.want)
			}
		})
	}
}
