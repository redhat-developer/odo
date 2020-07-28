package env

import (
	"reflect"
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
  Name: Use this value to set component name
  Namespace: Use this value to set component namespace
  DebugPort: Use this value to set component debug port`

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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got asd %s asd, want asd %s asd", got, tt.want)
			}
		})
	}
}

func TestIsSupportedParameter(t *testing.T) {
	supportedSetParameters := map[string]string{
		nameParameter:      nameParameterDescription,
		namespaceParameter: namespaceParameterDescription,
		debugportParameter: debugportParameterDescription,
	}

	tests := []struct {
		name                string
		parameter           string
		supportedParameters map[string]string
		want                bool
	}{
		{
			name:                "Case 1: Test supported parameter",
			parameter:           "Name",
			supportedParameters: supportedSetParameters,
			want:                true,
		},
		{
			name:                "Case 2: Test unsupported parameter",
			parameter:           "Fake",
			supportedParameters: supportedSetParameters,
			want:                false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSupportedParameter(tt.parameter, tt.supportedParameters)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got %t, want %t", got, tt.want)
			}
		})
	}
}
