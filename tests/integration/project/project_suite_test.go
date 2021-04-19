package project

import (
	"testing"

	"github.com/openshift/odo/tests/helper"
)

func TestProject(t *testing.T) {
	helper.RunTestSpecs(t, "Project Suite")
}
