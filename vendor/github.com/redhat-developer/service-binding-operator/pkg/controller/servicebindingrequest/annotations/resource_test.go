package annotations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiscoverBindingType(t *testing.T) {
	type args struct {
		value     string
		expected  bindingType
		wantedErr error
	}

	assertDiscoverBindingType := func(args args) func(t *testing.T) {
		return func(t *testing.T) {
			got, err := discoverBindingType(args.value)
			if args.wantedErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, args.expected, got)
			}
		}
	}

	t.Run("env/secret", assertDiscoverBindingType(args{
		value:    "binding:env:object:secret",
		expected: BindingTypeEnvVar,
	}))

	t.Run("env/configmap", assertDiscoverBindingType(args{
		value:    "binding:env:object:configmap",
		expected: BindingTypeEnvVar,
	}))

	t.Run("volumemount/secret", assertDiscoverBindingType(args{
		value:    "binding:volumemount:secret",
		expected: BindingTypeVolumeMount,
	}))

	t.Run("unknown/secret", assertDiscoverBindingType(args{
		value:     "binding:unknown:object:secret",
		expected:  BindingTypeEnvVar,
		wantedErr: UnknownBindingTypeErr("unknown"),
	}))
}
