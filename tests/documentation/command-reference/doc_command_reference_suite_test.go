//go:build !race
// +build !race

package docautomation

import (
	"testing"

	"github.com/redhat-developer/odo/tests/helper"
)

func TestDocCommandReferenceAutomation(t *testing.T) {
	helper.RunTestSpecs(t, "Doc Command Reference Automation Suite")
}
