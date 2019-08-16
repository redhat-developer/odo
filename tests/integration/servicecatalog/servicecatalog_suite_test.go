package integration

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestServicecatalog(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Servicecatalog Suite")
}
