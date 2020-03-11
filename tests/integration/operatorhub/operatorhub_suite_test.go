package integration_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOperatorhub(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Operatorhub Suite")
}
