//go:build !race
// +build !race

package docautomation

import (
	"testing"

	"github.com/redhat-developer/odo/tests/helper"
)

func TestDocAutomation(t *testing.T) {
	helper.RunTestSpecs(t, "Doc Automation Suite")
}
