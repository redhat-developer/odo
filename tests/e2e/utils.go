package e2e

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/gomega"
)

func createDir(dirName string) error {
	return os.MkdirAll(dirName, 0777)
}

func createFileAtPath(filePath string, fileName string) error {
	_, err := os.Stat(filePath + "/" + fileName)
	if os.IsNotExist(err) {
		file, err := os.Create(filePath + "/" + fileName)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	return nil
}

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

func generateTimeBasedName(prefix string) string {
	var t = strconv.FormatInt(time.Now().Unix(), 10)
	return fmt.Sprintf("%s-%s", prefix, t)
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

func retryingForOutputMatchStringOfHTTPResponse(url, match string, retry, sleep int) bool {
	for i := 0; i < retry; i++ {
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println(err.Error())
		}
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			body, _ := ioutil.ReadAll(resp.Body)
			//str := fmt.Sprintf("%s", body)
			if strings.Contains(string(body), match) {
				return true
			}
		}
		time.Sleep(time.Duration(sleep) * time.Second)
	}
	return false
}

func findLocalConfigValueOfGivenKey(key string) string {

	stdOut, _, _ := cmdRunner("odo utils config view")
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

func findGlobalConfigValueOfGivenKey(key string) string {
	stdOut, _, _ := cmdRunner("odo utils config view --global")
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

func getBuildNameUsingOc(compName string) string {
	stdOut, stdErr, _ := cmdRunner("oc get builds --output='name'")
	if stdErr != "" {
		return stdErr
	}
	re := regexp.MustCompile(compName + `-\S+`)
	buildName := re.FindString(stdOut)
	return strings.TrimSpace(buildName)
}

func getBuildParametersValueUsingOc(compName string) string {
	stdOut, stdErr, _ := cmdRunner("oc get builds")
	if stdErr != "" {
		return stdErr
	}
	re := regexp.MustCompile(compName + `-.+`)
	buildParametersValue := re.FindString(stdOut)
	return strings.TrimSpace(buildParametersValue)
}

func getDcNameUsingOc(compName string) string {
	stdOut, stdErr, _ := cmdRunner("oc get dc")
	if stdErr != "" {
		return stdErr
	}
	re := regexp.MustCompile(compName + `-\S+ `)
	dcName := re.FindString(stdOut)
	return strings.TrimSpace(dcName)
}

func getDcStatusValueUsingOc(compName string) string {
	stdOut, stdErr, _ := cmdRunner("oc get dc")
	if stdErr != "" {
		return stdErr
	}
	re := regexp.MustCompile(compName + `-\S+\s+[0-9]`)
	dcStatusCheckString := re.FindString(stdOut)
	return strings.TrimSpace(strings.SplitN(dcStatusCheckString, " ", 2)[1])
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
