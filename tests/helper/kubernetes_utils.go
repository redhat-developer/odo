package helper

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// CreateRandNamespace create new project with random name in kubernetes cluster (10 letters)
func CreateRandNamespace(context string) string {
	projectName := RandString(10)
	fmt.Fprintf(GinkgoWriter, "Creating a new project: %s\n", projectName)
	CmdShouldPass("kubectl", "create", "namespace", projectName)
	CmdShouldPass("kubectl", "config", "set-context", context, "--namespace", projectName)
	session := CmdShouldPass("kubectl", "get", "namespaces")
	Expect(session).To(ContainSubstring(projectName))
	return projectName
}

// DeleteNamespace deletes a specified project in kubernetes cluster
func DeleteNamespace(projectName string) {
	fmt.Fprintf(GinkgoWriter, "Deleting project: %s\n", projectName)
	CmdShouldPass("kubectl", "delete", "namespaces", projectName)
}
