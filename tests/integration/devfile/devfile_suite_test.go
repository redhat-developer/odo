package devfile

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/openshift/odo/tests/helper"
)

func TestDevfiles(t *testing.T) {
	helper.RunTestSpecs(t, "Devfile Suite")
}

var _ = BeforeSuite(func() {
	const registryName string = "DefaultDevfileRegistry"
	const addRegistryURL string = "https://registry.stage.devfile.io"

	// Use staging OCI-based registry for tests to avoid a potential overload
	helper.Cmd("odo", "registry", "update", registryName, addRegistryURL, "-f").ShouldPass()
})
