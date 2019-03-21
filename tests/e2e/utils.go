package e2e

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/gomega"
)

var ci = os.Getenv("CI")

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

// returns a local config value of given key or
// returns an empty string if value is not set
func getConfigValue(key string) string {
	stdOut, _, _ := cmdRunner("odo config view")
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

// returns a global config value of given key or
// returns an empty string if value is not set
func getPreferenceValue(key string) string {
	stdOut, _, _ := cmdRunner("odo preference view")
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

// replace and save a specified text with a given text from a file
// present in the path, returns error if unsuccessful
func replaceTextInFile(filePath string, actualString string, replaceString string) error {
	input, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	output := bytes.Replace(input, []byte(actualString), []byte(replaceString), 1)
	if err = ioutil.WriteFile(filePath, output, 0666); err != nil {
		return err
	}
	return nil
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

// This function executes oc command and returns the running pod name of a delopyed
// component by passing component name as a argument
func getRunningPodNameOfComp(compName string) string {
	stdOut, stdErr, _ := cmdRunner("oc get pods")
	if stdErr != "" {
		return stdErr
	}
	re := regexp.MustCompile(`(` + compName + `-\S+)\s+\S+\s+Running`)
	podName := re.FindStringSubmatch(stdOut)[1]
	return strings.TrimSpace(podName)
}

// This function execute oc command and returns build name of a delopyed
// component by passing component name as a argument
func getBuildName(compName string) string {
	stdOut, stdErr, _ := cmdRunner("oc get builds --output='name'")
	if stdErr != "" {
		return stdErr
	}
	re := regexp.MustCompile(compName + `-\S+`)
	buildName := re.FindString(stdOut)
	return strings.TrimSpace(buildName)
}

// This function execute oc command and returns parameter values of a delopyed
// component by passing component name as a argument
func getBuildParameterValues(compName string) string {
	stdOut, stdErr, _ := cmdRunner("oc get builds")
	if stdErr != "" {
		return stdErr
	}
	re := regexp.MustCompile(compName + `-.+`)
	buildParametersValue := re.FindString(stdOut)
	return strings.TrimSpace(buildParametersValue)
}

// This function execute oc command and returns dc name of a delopyed
// component by passing component name as a argument
func getDcName(compName string) string {
	stdOut, stdErr, _ := cmdRunner("oc get dc")
	if stdErr != "" {
		return stdErr
	}
	re := regexp.MustCompile(compName + `-\S+ `)
	dcName := re.FindString(stdOut)
	return strings.TrimSpace(dcName)
}

// This function execute oc command and returns dc REVISION
// status of a delopyed component by passing component name as a argument
func getDcStatusValue(compName string) string {
	stdOut, stdErr, _ := cmdRunner("oc get dc")
	if stdErr != "" {
		return stdErr
	}
	re := regexp.MustCompile(compName + `-\S+\s+[0-9]`)
	dcStatusCheckString := re.FindString(stdOut)
	return strings.TrimSpace(strings.SplitN(dcStatusCheckString, " ", 2)[1])
}
