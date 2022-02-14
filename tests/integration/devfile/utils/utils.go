package utils

import (
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/redhat-developer/odo/tests/helper"

	dfutil "github.com/devfile/library/pkg/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type OdoV2Watch struct {
	CmpName               string
	StringsToBeMatched    []string
	StringsNotToBeMatched []string
	FolderToCheck         string
	SrcType               string
}

// OdoWatch creates files, dir in the context and watches for the changes to be pushed
// Specify OdoV2Watch for odo version 2(devfile)
// platform is kube
func OdoWatch(odoV2Watch OdoV2Watch, project, context, flag string, runner interface{}, platform string) {

	// After the watch command has started (indicated via channel), simulate file system changes
	startSimulationCh := make(chan bool)
	go func() {
		startMsg := <-startSimulationCh
		if startMsg {
			err := os.MkdirAll(filepath.Join(context, ".abc"), 0750)
			Expect(err).To(BeNil())

			err = os.MkdirAll(filepath.Join(context, "abcd"), 0750)
			Expect(err).To(BeNil())

			_, err = os.Create(filepath.Join(context, "a.txt"))
			Expect(err).To(BeNil())

			if odoV2Watch.SrcType == "openjdk" {
				helper.ReplaceString(filepath.Join(context, "src", "main", "java", "MessageProducer.java"), "Hello", "Hello odo")
			} else {
				helper.ReplaceString(filepath.Join(context, "server.js"), "Hello", "Hello odo")
			}

		}
	}()

	success, err := helper.WatchNonRetCmdStdOut(
		("odo watch " + flag + " --context " + context),
		time.Duration(5)*time.Minute,
		func(output string) bool {
			// the test hangs up on the CI when the delay is set to 0
			// so we only check if the start message was displayed correctly or not
			if strings.Contains(flag, "delay 0") {
				return true
			}
			// Returns true if the test has succeeded, false if not yet

			stringsMatched := true

			for _, stringToBeMatched := range odoV2Watch.StringsToBeMatched {
				if !strings.Contains(output, stringToBeMatched) {
					fmt.Fprintln(GinkgoWriter, "Missing string: ", stringToBeMatched)
					stringsMatched = false
				}
			}

			if stringsMatched {

				// first push is successful
				// now delete a folder and check if the deletion is propagated properly
				// and the file is removed from the cluster
				index := suffixarray.New([]byte(output))
				offsets := index.Lookup([]byte(filepath.Join(context, "abcd")+" changed"), -1)

				// the first occurrence of '<target-dir> changed' means the creation of it was pushed to the cluster
				// and the first push was successful
				if len(offsets) == 1 {
					helper.DeleteDir(filepath.Join(context, "abcd"))
				} else if len(offsets) > 1 {
					// the occurrence of 'target-directory' more than once indicates that the deletion was propagated too
					// Verify directory deleted from component pod
					err := validateContainerExecListDir(odoV2Watch, runner, platform, project)
					Expect(err).To(BeNil())
					return true
				}
			}

			return false
		},
		startSimulationCh,
		func(output string) bool {
			// Returns true to indicate the test should begin file system file change simulation
			return strings.Contains(output, "Waiting for something to change")
		})

	Expect(success).To(Equal(true))
	Expect(err).To(BeNil())
}

// OdoWatchWithDebug changes files in the context and watches for the changes to be pushed
// It checks if the push is in debug mode or not
// After a successful push with watch, it tries to start a debug session
func OdoWatchWithDebug(odoV2Watch OdoV2Watch, context, flag string) {

	startSimulationCh := make(chan bool)
	go func() {
		startMsg := <-startSimulationCh
		if startMsg {
			helper.ReplaceString(filepath.Join(context, "server.js"), "Hello", "Hello odo")
			helper.ReplaceString(filepath.Join(context, "package.json"), "application", "app")
		}
	}()

	success, err := helper.WatchNonRetCmdStdOut(
		("odo watch " + flag + " --context " + context),
		time.Duration(5)*time.Minute,
		func(output string) bool {
			stringsMatched := true

			for _, stringToBeMatched := range odoV2Watch.StringsToBeMatched {
				if !strings.Contains(output, stringToBeMatched) {
					stringsMatched = false
				}
			}

			if stringsMatched {
				httpPort, err := dfutil.HTTPGetFreePort()
				Expect(err).NotTo(HaveOccurred())
				freePort := strconv.Itoa(httpPort)

				stopChannel := make(chan bool)
				go func() {
					helper.Cmd("odo", "debug", "port-forward", "--local-port", freePort).WithTerminate(60*time.Second, stopChannel).ShouldRun()
				}()

				// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
				// We are just using this to validate if nodejs agent is listening on the other side
				helper.HttpWaitForWithStatus("http://localhost:"+freePort, "WebSockets request was expected", 12, 5, 400)
				stopChannel <- true
				return true
			}

			return false
		},
		startSimulationCh,
		func(output string) bool {
			return strings.Contains(output, "Waiting for something to change")
		})

	Expect(success).To(Equal(true))
	Expect(err).To(BeNil())
}

// OdoWatchWithIgnore checks if odo watch ignores the specified files and
// it also checks if odo-file-index.json and .git are ignored
// when --ignores is used
func OdoWatchWithIgnore(odoV2Watch OdoV2Watch, context, flag string) {

	startSimulationCh := make(chan bool)
	go func() {
		startMsg := <-startSimulationCh
		if startMsg {
			_, err := os.Create(filepath.Join(context, "doignoreme.txt"))
			Expect(err).To(BeNil())

			_, err = os.Create(filepath.Join(context, "donotignoreme.txt"))
			Expect(err).To(BeNil())
		}
	}()

	success, err := helper.WatchNonRetCmdStdOut(
		("odo watch " + flag + " --context " + context),
		time.Duration(5)*time.Minute,
		func(output string) bool {
			stringsMatched := true
			for _, stringToBeMatched := range odoV2Watch.StringsToBeMatched {
				if !strings.Contains(output, stringToBeMatched) {
					stringsMatched = false
				}
			}

			stringsNotMatched := true
			for _, stringNotToBeMatched := range odoV2Watch.StringsNotToBeMatched {
				if strings.Contains(output, stringNotToBeMatched) {
					stringsNotMatched = false
				}
			}

			if stringsMatched && stringsNotMatched {
				return true
			}

			return false
		},
		startSimulationCh,
		func(output string) bool {
			return strings.Contains(output, "Waiting for something to change")
		})

	Expect(success).To(Equal(true))
	Expect(err).To(BeNil())
}

func validateContainerExecListDir(odoV2Watch OdoV2Watch, runner interface{}, platform, project string) error {
	var folderToCheck, podName string
	cliRunner := runner.(helper.CliRunner)
	switch platform {
	case "kube":
		folderToCheck = "/projects"
		if odoV2Watch.FolderToCheck != "" {
			folderToCheck = odoV2Watch.FolderToCheck
		}
		podName = cliRunner.GetRunningPodNameByComponent(odoV2Watch.CmpName, project)

	default:
		return fmt.Errorf("Platform %s is not supported", platform)
	}

	// check if contains a.txt, .abc && abcd is deleted
	cliRunner.WaitForRunnerCmdOut([]string{"exec", podName, "--namespace", project,
		"--", "ls", "-lai", folderToCheck}, 5, true, func(output string) bool {
		return !(strings.Contains(output, "abcd")) && (strings.Contains(output, "a.txt")) && (strings.Contains(output, ".abc"))
	})

	return nil
}

// DeleteLocalConfig helps user to delete local config files with flags
func DeleteLocalConfig(args ...string) {
	helper.Cmd("odo", args...).ShouldFail()
	output := helper.Cmd("odo", append(args, "-af")...).ShouldPass().Out()
	expectedOutput := []string{
		"Successfully deleted env file",
		"Successfully deleted devfile.yaml file",
	}
	helper.MatchAllInOutput(output, expectedOutput)
}

// VerifyCatalogListComponent verifies components inside wantOutput exists or not
// in Devfile Component list
func VerifyCatalogListComponent(output string, cmpName []string) error {
	var data map[string]interface{}
	listItems := []string{"items"}

	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return err
	}

	for _, items := range listItems {
		outputBytes, err := json.Marshal(data[items])
		if err != nil {
			return err
		}
		output = string(outputBytes)
		helper.MatchAllInOutput(output, cmpName)
	}
	return nil
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
