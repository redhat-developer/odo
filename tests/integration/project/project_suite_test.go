package project

import (
	"testing"

	"github.com/redhat-developer/odo/tests/helper"
)

func TestProject(t *testing.T) {
	helper.RunTestSpecs(t, "Project Suite")
}
