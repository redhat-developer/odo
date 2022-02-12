package helper

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	dfutil "github.com/devfile/library/pkg/util"
)

// copyKubeConfigFile copies default kubeconfig file into current temporary context config file
func copyKubeConfigFile(kubeConfigFile, tempConfigFile string) {
	info, err := os.Stat(kubeConfigFile)
	Expect(err).NotTo(HaveOccurred())
	err = dfutil.CopyFile(kubeConfigFile, tempConfigFile, info)
	Expect(err).NotTo(HaveOccurred())
	os.Setenv("KUBECONFIG", tempConfigFile)
	fmt.Fprintf(GinkgoWriter, "Setting KUBECONFIG=%s\n", tempConfigFile)
}
