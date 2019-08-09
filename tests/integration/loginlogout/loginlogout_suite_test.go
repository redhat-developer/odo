package integration

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLoginlogout(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Loginlogout Suite")
}
