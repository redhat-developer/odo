package debug

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/openshift/odo/tests/helper"
)

func TestDebug(t *testing.T) {
	helper.RunTestSpecs(t, "Devfile Debug Suite")
}

// Use JustBeforeEach to use the preference file set into BeforeEach
var _ = JustBeforeEach(func() {
	helper.SetDefaultDevfileRegistryAsStaging()
})
