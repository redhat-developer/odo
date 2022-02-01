package helper

import (
	"fmt"
	"regexp"
	"strings"

	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// GetConfigValue returns a local config value of given key or
// returns an empty string if value is not set
func GetConfigValue(key string) string {
	return GetConfigValueWithContext(key, "")
}

// GetConfigValueWithContext returns a local config value of given key and contextdir or
// returns an empty string if value is not set
func GetConfigValueWithContext(key string, context string) string {
	var stdOut string
	if context != "" {
		stdOut = Cmd("odo", "config", "view", "--context", context).ShouldPass().Out()
	} else {
		stdOut = Cmd("odo", "config", "view").ShouldPass().Out()
	}
	re := regexp.MustCompile(key + `.+`)
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

// GetLocalEnvInfoValueWithContext returns an envInfo value of given key and contextdir or
// returns an empty string if value is not set
func GetLocalEnvInfoValueWithContext(key string, context string) string {
	var stdOut string
	if context != "" {
		stdOut = Cmd("odo", "env", "view", "--context", context).ShouldPass().Out()
	} else {
		stdOut = Cmd("odo", "env", "view").ShouldPass().Out()
	}
	re := regexp.MustCompile(key + `.+`)
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

// GetPreferenceValue a global config value of given key or
// returns an empty string if value is not set
func GetPreferenceValue(key string) string {
	stdOut := Cmd("odo", "preference", "view").ShouldPass().Out()
	re := regexp.MustCompile(key + `.+`)
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

// DetermineRouteURL takes context path as argument and returns the http URL
// where the current component exposes it's service this URL can
// then be used in order to interact with the deployed service running in Openshift
func DetermineRouteURL(context string) string {
	urls := DetermineRouteURLs(context)
	// only return the 1st element if it exists
	if len(urls) > 0 {
		return urls[0]
	}

	return ""
}

// DetermineRouteURLs takes context path as argument and returns the URLs
// where the current component exposes it's service, these URLs can
// then be used in order to interact with the deployed service running in Openshift
func DetermineRouteURLs(context string) []string {
	var stdOut string
	if context != "" {
		stdOut = Cmd("odo", "url", "list", "--context", context).ShouldPass().Out()
	} else {
		stdOut = Cmd("odo", "url", "list").ShouldPass().Out()
	}
	reURL := regexp.MustCompile(`\s+http(s?)://.\S+`)
	odoURLs := reURL.FindAllString(stdOut, -1)
	for i := range odoURLs {
		odoURLs[i] = strings.TrimSpace(odoURLs[i])
	}
	return odoURLs
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
