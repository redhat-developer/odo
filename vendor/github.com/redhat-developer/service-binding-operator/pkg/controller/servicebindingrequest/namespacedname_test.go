package servicebindingrequest

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
)

func TestNamespacednameIsSBRNamespacedNameEmpty(t *testing.T) {
	require.True(t, IsNamespacedNameEmpty(types.NamespacedName{}))
	require.True(t, IsNamespacedNameEmpty(types.NamespacedName{Namespace: "ns"}))
	require.True(t, IsNamespacedNameEmpty(types.NamespacedName{Name: "name"}))
	require.False(t, IsNamespacedNameEmpty(types.NamespacedName{Namespace: "ns", Name: "name"}))
}
