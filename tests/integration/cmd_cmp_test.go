// +build !race

package integration

import (
	. "github.com/onsi/ginkgo"
)

func componentTestsNoSub() {
	componentTests()
}

var _ = Describe("odo component command tests", componentTestsNoSub)
