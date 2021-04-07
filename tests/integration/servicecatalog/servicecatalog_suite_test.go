package integration

import (
	"testing"

	"github.com/openshift/odo/tests/helper"
)

func TestServicecatalog(t *testing.T) {
	helper.RunTestSpecs(t, "Servicecatalog Suite")
}
