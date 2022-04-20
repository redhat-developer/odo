package utils

import (
	"strings"

	"github.com/redhat-developer/odo/tests/helper"

	. "github.com/onsi/gomega"
)

type OdoV2Watch struct {
	CmpName               string
	StringsToBeMatched    []string
	StringsNotToBeMatched []string
	FolderToCheck         string
	SrcType               string
}

// VerifyContainerSyncEnv verifies the sync env in the container
func VerifyContainerSyncEnv(podName, containerName, namespace, projectSourceValue, projectsRootValue string, cliRunner helper.CliRunner) {
	envProjectsRoot, envProjectSource := "PROJECTS_ROOT", "PROJECT_SOURCE"
	projectSourceMatched, projectsRootMatched := false, false

	envNamesAndValues := cliRunner.GetContainerEnv(podName, "runtime", namespace)
	envNamesAndValuesArr := strings.Fields(envNamesAndValues)

	for _, envNamesAndValues := range envNamesAndValuesArr {
		envNameAndValueArr := strings.Split(envNamesAndValues, ":")

		if envNameAndValueArr[0] == envProjectSource && strings.Contains(envNameAndValueArr[1], projectSourceValue) {
			projectSourceMatched = true
		}

		if envNameAndValueArr[0] == envProjectsRoot && strings.Contains(envNameAndValueArr[1], projectsRootValue) {
			projectsRootMatched = true
		}
	}

	Expect(projectSourceMatched).To(Equal(true))
	Expect(projectsRootMatched).To(Equal(true))
}
