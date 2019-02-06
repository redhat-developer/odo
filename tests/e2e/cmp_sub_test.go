// +build !race

package e2e

import (
	. "github.com/onsi/ginkgo"
)

func componentTestsSub() {
	componentTests("odo component")
}

var _ = Describe("odoCmpSubE2e", componentTestsSub)
