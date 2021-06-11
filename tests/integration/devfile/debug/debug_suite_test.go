package debug

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/openshift/odo/tests/helper"
)

func TestDebug(t *testing.T) {
	helper.RunTestSpecs(t, "Devfile Debug Suite")
}

var _ = BeforeSuite(func() {
	const registryName string = "DefaultDevfileRegistry"
	const addRegistryURL string = "https://registry.stage.devfile.io"

	// Use staging OCI-based registry for tests to avoid a potential overload
	helper.CmdShouldPass("odo", "registry", "update", registryName, addRegistryURL)
})
