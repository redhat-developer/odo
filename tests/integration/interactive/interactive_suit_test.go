//go:build !race
// +build !race

package interactive

import (
	"testing"

	"github.com/redhat-developer/odo/tests/helper"
)

func TestInteractive(t *testing.T) {
	helper.RunTestSpecs(t, "Interactive Suite")
}
