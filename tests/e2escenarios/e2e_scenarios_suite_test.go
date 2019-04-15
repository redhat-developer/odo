package e2escenarios

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestE2eScenarios(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "odo e2e scenarios")
}
