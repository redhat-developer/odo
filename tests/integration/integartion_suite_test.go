// +build !race

package integration

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegartion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integartion Suite")
}
