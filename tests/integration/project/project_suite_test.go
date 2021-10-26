package project

import (
	"testing"

	"github.com/openshift/odo/v2/tests/helper"
)

func TestProject(t *testing.T) {
	helper.RunTestSpecs(t, "Project Suite")
}
