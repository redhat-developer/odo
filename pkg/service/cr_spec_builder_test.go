package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCRSpecBuilderSetAndValidate(t *testing.T) {
	builder := NewCRSpecBuilder(MockCRDescriptionOne().SpecDescriptors)
	err := builder.SetAndValidate("size", "3")
	require.Nil(t, err, "set shouldn't fail")
	result := builder.JSON()
	require.Equal(t, result, `{"size":3}`)
	// second time to confirm it doesn't duplicate
	err = builder.SetAndValidate("size", "3")
	require.Nil(t, err, "set shouldn't fail")
	result = builder.JSON()
	require.Equal(t, result, `{"size":3}`)
	// incorrect argument
	err = builder.SetAndValidate("seze", "3")
	require.NotNil(t, err, "set should fail")
	require.Equal(t, err.Error(), fmt.Sprintf("the parameter %s is not present in the Operand Schema", "seze"), "error statement should match")
}
