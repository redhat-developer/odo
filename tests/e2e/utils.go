package e2e

import (
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/gomega"
)

// createFileAtPath creates a file at the given path and writes the given content
// path is the path to the required file
// fileContent is the content to be written to the given file
func createFileAtPathWithContent(path string, fileContent string) error {
	// check if file exists
	var _, err = os.Stat(path)

	var file *os.File

	// create file if not exists
	if os.IsNotExist(err) {
		file, err = os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()

	} else {
		// Open file using READ & WRITE permission.
		file, err = os.OpenFile(path, os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	// write to file
	_, err = file.WriteString(fileContent)
	if err != nil {
		return err
	}

	return nil
}

// determineRouteURL returns the http URL where the current component exposes it's service
// this URL can then be used in order to interact with the deployed service running in Openshift
// keeping with the spirit of the e2e tests, this expects, odo, sed and awk to be on the PATH
func determineRouteURL() string {
	output := runCmdShouldPass("odo url list  | sed -n '1!p' | awk 'FNR==2 { print $2 }'")
	return strings.TrimSpace(output)
}

func odoCreateProject(projectName string) {
	runCmdShouldPass("odo project create " + projectName)
	waitForCmdOut("odo project set "+projectName, 4, false, func(output string) bool {
		return strings.Contains(output, "Already on project : "+projectName)
	})
}

// deletes a specified project
func odoDeleteProject(project string) {
	var waitOut bool
	if len(project) > 0 {
		waitOut = waitForCmdOut(fmt.Sprintf("odo project delete -f %s", project), 10, true, func(out string) bool {
			return strings.Contains(out, fmt.Sprintf("Deleted project : %s", project))
		})
		Expect(waitOut).To(BeTrue())
		waitForDeleteCmd("odo project list", project)
	}
}

// cleanUpAfterProjects cleans up projects, after deleting them
func cleanUpAfterProjects(projects []string) {
	for _, p := range projects {
		odoDeleteProject(p)
	}
}
