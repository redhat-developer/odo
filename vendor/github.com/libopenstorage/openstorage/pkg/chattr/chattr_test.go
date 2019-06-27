package chattr

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testFile = "/tmp/osd-test"
)

func TestImmutable(t *testing.T) {
	// create a test file
	_, err := os.Create(testFile)
	require.NoError(t, err, "Unexpected error on create test file")

	// chattr +i
	err = AddImmutable(testFile)
	require.NoError(t, err, "Unexpected error on AddImmutable")

	// check if +i is added on the file
	isImmutable := IsImmutable(testFile)
	require.True(t, isImmutable, "Unexpected: Path is not immutable")

	// remove should fail
	err = os.RemoveAll(testFile)
	require.Error(t, err, "Expected an error on remove")

	// check if file still exists
	_, err = os.Stat(testFile)
	require.NoError(t, err, "Expected the file to be present")

	// chattr -i
	err = RemoveImmutable(testFile)
	require.NoError(t, err, "Unexpected error on RemoveImmutable")

	// check if +i is removed on the file
	isImmutable = IsImmutable(testFile)
	require.False(t, isImmutable, "Unexpected: Path is not mutable")

	// remove should succeed
	err = os.RemoveAll(testFile)
	require.NoError(t, err, "Unexpected an error on remove")

	// check if file still exists
	_, err = os.Stat(testFile)
	require.Error(t, err, "Expected the file to be removed")
	require.True(t, os.IsNotExist(err), "Unexpected error on remove")

}
