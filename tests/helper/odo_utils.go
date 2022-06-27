package helper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"

	"github.com/redhat-developer/odo/pkg/devfile"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// GetPreferenceValue a global config value of given key or
// returns an empty string if value is not set
func GetPreferenceValue(key string) string {
	stdOut := Cmd("odo", "preference", "view").ShouldPass().Out()
	re := regexp.MustCompile(" " + key + `.+`)
	odoConfigKeyValue := re.FindString(stdOut)
	if odoConfigKeyValue == "" {
		return fmt.Sprintf("%s not found", key)
	}
	trimKeyValue := strings.TrimSpace(odoConfigKeyValue)
	if strings.Compare(key, trimKeyValue) != 0 {
		return strings.TrimSpace(strings.SplitN(trimKeyValue, " ", 2)[1])
	}
	return ""
}

// CreateRandProject create new project with random name (10 letters)
// without writing to the config file (without switching project)
func CreateRandProject() string {
	projectName := SetProjectName()
	fmt.Fprintf(GinkgoWriter, "Creating a new project: %s\n", projectName)
	session := Cmd("odo", "project", "create", projectName, "-w", "-v4").ShouldPass().Out()
	Expect(session).To(ContainSubstring("New project created"))
	Expect(session).To(ContainSubstring(projectName))
	return projectName
}

// DeleteProject deletes a specified project
func DeleteProject(projectName string) {
	fmt.Fprintf(GinkgoWriter, "Deleting project: %s\n", projectName)
	session := Cmd("odo", "project", "delete", projectName, "-f").ShouldPass().Out()
	Expect(session).To(ContainSubstring("Deleted project : " + projectName))
}

// GetMetadataFromDevfile retrieves the metadata from devfile
func GetMetadataFromDevfile(devfilePath string) devfilepkg.DevfileMetadata {
	devObj, err := devfile.ParseAndValidateFromFile(devfilePath)
	Expect(err).ToNot(HaveOccurred())
	return devObj.Data.GetMetadata()
}

func GetDevfileComponents(devfilePath, componentName string) []v1alpha2.Component {
	devObj, err := devfile.ParseAndValidateFromFile(devfilePath)
	Expect(err).ToNot(HaveOccurred())
	components, err := devObj.Data.GetComponents(common.DevfileOptions{
		FilterByName: componentName,
	})
	Expect(err).ToNot(HaveOccurred())
	return components
}
