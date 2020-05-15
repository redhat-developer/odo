package helper

import (
	"os"

	. "github.com/onsi/gomega"
)

// copyKubeConfigFile copies default kubeconfig file into current temporary context config file
func copyKubeConfigFile(kubeConfigFile, tempConfigFile string) {
	info, err := os.Stat(kubeConfigFile)
	Expect(err).NotTo(HaveOccurred())
	err = copyFile(kubeConfigFile, tempConfigFile, info)
	Expect(err).NotTo(HaveOccurred())
	os.Setenv("KUBECONFIG", tempConfigFile)
}
