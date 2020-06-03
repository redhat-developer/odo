package helper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/pkg/util"
)

// copyKubeConfigFile copies default kubeconfig file into current temporary context config file
func copyKubeConfigFile(kubeConfigFile, tempConfigFile string) {
	info, err := os.Stat(kubeConfigFile)
	Expect(err).NotTo(HaveOccurred())
	err = util.CopyFile(kubeConfigFile, tempConfigFile, info)
	Expect(err).NotTo(HaveOccurred())
	os.Setenv("KUBECONFIG", tempConfigFile)
	fmt.Fprintf(GinkgoWriter, "Setting KUBECONFIG=%s\n", tempConfigFile)
}

// GetExistingKubeConfigPath retrieves the Kubernetes configuration from the most appropriate location
func GetExistingKubeConfigPath() string {

	// 1) If KUBECONFIG env var is specified, return that path
	kubeconfigEnv := strings.TrimSpace(os.Getenv("KUBECONFIG"))
	if len(kubeconfigEnv) != 0 {
		return kubeconfigEnv
	}

	// 2) Otherwise return the default config path
	homeDir := GetUserHomeDir()
	return filepath.Join(homeDir, ".kube", "config")

}
