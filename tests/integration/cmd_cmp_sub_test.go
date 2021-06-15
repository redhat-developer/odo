// +build !race

package integration

import (
	. "github.com/onsi/ginkgo"
)

func componentTestsSub() {
	//componentTests("component")
}

var _ = Describe("odo sub component command tests", componentTestsSub)
