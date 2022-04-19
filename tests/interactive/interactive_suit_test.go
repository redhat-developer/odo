//go:build !race && (linux || darwin || dragonfly || solaris || openbsd || netbsd || freebsd)
// +build !race
// +build linux darwin dragonfly solaris openbsd netbsd freebsd

package interactive

import (
	"testing"

	"github.com/redhat-developer/odo/tests/helper"
)

func TestInteractive(t *testing.T) {
	helper.RunTestSpecs(t, "Interactive Suite")
}
