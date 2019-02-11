// +build !race

package e2e

import (
	. "github.com/onsi/ginkgo"
)

func componentTestsNoSub() {
	componentTests("odo")
}

var _ = Describe("odoCmpE2e", componentTestsNoSub)
