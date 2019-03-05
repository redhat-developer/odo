package e2e

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

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
func determineRouteURL() string {
	stdOut, stdErr, exitCode := cmdRunner("odo url list")
	if exitCode != 0 {
		return stdErr
	}
	reURL := regexp.MustCompile(`\s+http://.\S+`)
	odoURL := reURL.FindString(stdOut)
	return strings.TrimSpace(odoURL)
}

// creates the specified namespace
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
	}
}

// cleanUpAfterProjects cleans up projects, after deleting them
func cleanUpAfterProjects(projects []string) {
	for _, p := range projects {
		odoDeleteProject(p)
	}
}

// getActiveApplication returns the active application in the project
// returns empty string if no active application is present in the project
func getActiveApplication() string {
	stdOut, stdErr, exitCode := cmdRunner("odo app list")
	if exitCode != 0 {
		return stdErr
	}
	if strings.Contains(strings.ToLower(stdOut), "no applications") {
		return ""
	}
	reActiveApp := regexp.MustCompile(`[*]\s+\S+`)
	odoActiveApp := strings.Split(reActiveApp.FindString(stdOut), "*")[1]
	return strings.TrimSpace(odoActiveApp)
}

// This function keeps trying in a regular interval of time to find a given string
// match for a perticular timeout period against a http response. returns true
// if string matches and response status code is 200, returns false otherwise
// It takes 4 arguments
// url - HTTP(S) URL (string)
// match - Sub string you are looking for from the response (string)
// retry - No of retry to fing the match string (int)
// sleep - Time interval of each try (int)
func matchResponseSubString(url, match string, retry, sleep int) bool {
	var i int
	for i := 0; i < retry; i++ {
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println(err.Error())
		}
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			body, _ := ioutil.ReadAll(resp.Body)
			if strings.Contains(string(body), match) {
				return true
			}
		}
		time.Sleep(time.Duration(sleep) * time.Second)
	}
	fmt.Printf("Could not get the match string \"%s\" in %d seconds\n", match, i)
	return false
}
