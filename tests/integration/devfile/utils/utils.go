package utils

import (
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/odo/pkg/util"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func useProjectIfAvailable(args []string, project string) []string {
	if project != "" {
		args = append(args, "--project", project)
	}

	return args
}

// ExecDefaultDevfileCommands executes the default devfile commands
func ExecDefaultDevfileCommands(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "java-springboot", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()

	helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.Cmd("odo", args...).ShouldPass().Out()
	helper.MatchAllInOutput(output, []string{
		"Executing defaultbuild command",
		"mvn clean",
		"Executing defaultrun command",
		"spring-boot:run",
	})
}

// ExecWithMissingBuildCommand executes odo push with a missing build command
func ExecWithMissingBuildCommand(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-without-devbuild.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.Cmd("odo", args...).ShouldPass().Out()
	Expect(output).NotTo(ContainSubstring("Executing devbuild command"))
	Expect(output).To(ContainSubstring("Executing devrun command \"npm install && npm start\""))
}

// ExecWithMissingRunCommand executes odo push with a missing run command
func ExecWithMissingRunCommand(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	// Remove the run commands
	helper.ReplaceString(filepath.Join(projectDirPath, "devfile.yaml"), "kind: run", "kind: debug")

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.Cmd("odo", args...).ShouldFail().Err()
	Expect(output).NotTo(ContainSubstring("Executing devrun command"))
	Expect(output).To(ContainSubstring("the command group of kind \"run\" is not found in the devfile"))
}

// ExecWithCustomCommand executes odo push with a custom command
func ExecWithCustomCommand(projectDirPath, cmpName, namespace string) {
}

// ExecWithWrongCustomCommand executes odo push with a wrong custom command
func ExecWithWrongCustomCommand(projectDirPath, cmpName, namespace string) {
}

// ExecWithMultipleOrNoDefaults executes odo push with multiple or no default commands
func ExecWithMultipleOrNoDefaults(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-multiple-defaults.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.Cmd("odo", args...).ShouldFail().Err()
	helper.MatchAllInOutput(output, []string{
		"group test error",
		"currently there is more than one default command",
	})

	helper.DeleteFile(filepath.Join(projectDirPath, "devfile.yaml"))
	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-no-default.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	output = helper.Cmd("odo", args...).ShouldFail().Err()
	helper.MatchAllInOutput(output, []string{
		"group run error",
		"currently there is no default command",
	})
}

// ExecCommandWithoutGroupUsingFlags executes odo push with no command group using flags
func ExecCommandWithoutGroupUsingFlags(projectDirPath, cmpName, namespace string) {
}

// ExecWithInvalidCommandGroup executes odo push with an invalid command group
func ExecWithInvalidCommandGroup(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "java-springboot", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()

	helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	// Remove the run commands
	helper.ReplaceString(filepath.Join(projectDirPath, "devfile.yaml"), "kind: build", "kind: init")

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.Cmd("odo", args...).ShouldFail().Err()
	Expect(output).To(ContainSubstring("must be one of the following: \"build\", \"run\", \"test\", \"debug\""))
}

func ExecPushToTestParent(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-parent.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()

	args = append(args, "--build-command", "devbuild", "-f")
	output := helper.Cmd("odo", args...).ShouldPass().Out()
	helper.MatchAllInOutput(output, []string{"Executing devbuild command", "touch blah.js"})
}

func ExecPushWithParentOverride(projectDirPath, cmpName, appName, namespace string, freePort int) {
	args := []string{"create", "nodejs", cmpName, "--app", appName}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "parentSupport", "devfile-with-parent.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	// update the devfile with the free port
	helper.ReplaceString(filepath.Join(projectDirPath, "devfile.yaml"), "(-1)", strconv.Itoa(freePort))

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()
}

func ExecPushWithCompositeOverride(projectDirPath, cmpName, namespace string, freePort int) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "parentSupport", "devfile-with-parent-composite.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	// update the devfile with the free port
	helper.ReplaceString(filepath.Join(projectDirPath, "devfile.yaml"), "(-1)", strconv.Itoa(freePort))

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.Cmd("odo", args...).ShouldPass().Out()

	helper.MatchAllInOutput(output, []string{"Executing createfile command", "touch /projects/testfile"})
}

func ExecPushWithMultiLayerParent(projectDirPath, cmpName, appName, namespace string, freePort int) {
	args := []string{"create", "nodejs", cmpName, "--app", appName}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "parentSupport", "devfile-with-multi-layer-parent.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	// update the devfile with the free port
	helper.ReplaceString(filepath.Join(projectDirPath, "devfile.yaml"), "(-1)", strconv.Itoa(freePort))

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()

	args = []string{"push", "--build-command", "devbuild", "-f"}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()

	args = []string{"push", "--build-command", "build", "-f"}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()
}

// ExecPushToTestFileChanges executes odo push with and without a file change
func ExecPushToTestFileChanges(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()

	output := helper.Cmd("odo", args...).ShouldPass().Out()
	Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

	helper.ReplaceString(filepath.Join(projectDirPath, "server.js"), "Hello from Node.js", "UPDATED!")
	output = helper.Cmd("odo", args...).ShouldPass().Out()
	Expect(output).To(ContainSubstring("Syncing files to the component"))
}

// ExecPushWithForceFlag executes odo push with a force flag
func ExecPushWithForceFlag(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()

	// use the force build flag and push
	args = []string{"push", "-f"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.Cmd("odo", args...).ShouldPass().Out()
	Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))
}

// ExecPushWithNewFileAndDir executes odo push after creating a new file and dir
func ExecPushWithNewFileAndDir(projectDirPath, cmpName, namespace, newFilePath, newDirPath string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	// Create a new file that we plan on deleting later...
	if err := helper.CreateFileWithContent(newFilePath, "hello world"); err != nil {
		fmt.Printf("the foobar.txt file was not created, reason %v", err.Error())
	}

	// Create a new directory
	helper.MakeDir(newDirPath)

	// Push
	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	helper.Cmd("odo", args...).ShouldPass()
}

// ExecWithHotReload executes odo push with hot reload true
func ExecWithHotReload(projectDirPath, cmpName, namespace string, hotReload bool) {
}

type OdoV1Watch struct {
	SrcType  string
	RouteURL string
	AppName  string
}

type OdoV2Watch struct {
	CmpName               string
	StringsToBeMatched    []string
	StringsNotToBeMatched []string
	FolderToCheck         string
	SrcType               string
}

// OdoWatch creates files, dir in the context and watches for the changes to be pushed
// Specify OdoV1Watch for odo version 1, OdoV2Watch for odo version 2(devfile)
// platform is kube
func OdoWatch(odoV1Watch OdoV1Watch, odoV2Watch OdoV2Watch, project, context, flag string, runner interface{}, platform string) {

	isDevfileTest := false

	// if the odoV2Watch object is not empty, its a devfile test
	if !reflect.DeepEqual(odoV2Watch, OdoV2Watch{}) {
		isDevfileTest = true
	}

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

			if isDevfileTest {
				if odoV2Watch.SrcType == "openjdk" {
					helper.ReplaceString(filepath.Join(context, "src", "main", "java", "MessageProducer.java"), "Hello", "Hello odo")
				} else {
					helper.ReplaceString(filepath.Join(context, "server.js"), "Hello", "Hello odo")
				}
			} else {
				helper.DeleteDir(filepath.Join(context, "abcd"))
				if odoV1Watch.SrcType == "openjdk" {
					helper.ReplaceString(filepath.Join(context, "src", "main", "java", "MessageProducer.java"), "Hello", "Hello odo")
				} else {
					helper.ReplaceString(filepath.Join(context, "server.js"), "Hello", "Hello odo")
				}
			}
		}
	}()

	if !isDevfileTest {
		flag = strings.TrimSpace(fmt.Sprintf("-v 4 %s", flag))
	}

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
			if isDevfileTest {
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
						err := validateContainerExecListDir(odoV1Watch, odoV2Watch, runner, platform, project, isDevfileTest)
						Expect(err).To(BeNil())
						return true
					}
				}
			} else {
				curlURL := helper.Cmd("curl", odoV1Watch.RouteURL).ShouldPass().Out()
				if strings.Contains(curlURL, "Hello odo") {
					// Verify delete from component pod
					err := validateContainerExecListDir(odoV1Watch, odoV2Watch, runner, platform, project, isDevfileTest)
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

	if !isDevfileTest {
		// Verify memory limits to be same as configured
		getMemoryLimit := runner.(helper.OcRunner).MaxMemory(odoV1Watch.SrcType+"-app", odoV1Watch.AppName, project)
		Expect(getMemoryLimit).To(ContainSubstring("700Mi"))
		getMemoryRequest := runner.(helper.OcRunner).MinMemory(odoV1Watch.SrcType+"-app", odoV1Watch.AppName, project)
		Expect(getMemoryRequest).To(ContainSubstring("400Mi"))
	}
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
				httpPort, err := util.HTTPGetFreePort()
				Expect(err).NotTo(HaveOccurred())
				freePort := strconv.Itoa(httpPort)

				stopChannel := make(chan bool)
				go func() {
					helper.Cmd("odo", "debug", "port-forward", "--local-port", freePort).WithTerminate(60*time.Second, stopChannel).ShouldRun()
				}()

				// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
				// We are just using this to validate if nodejs agent is listening on the other side
				helper.HttpWaitForWithStatus("http://localhost:"+freePort, "WebSockets request was expected", 12, 5, 400)

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

func validateContainerExecListDir(odoV1Watch OdoV1Watch, odoV2Watch OdoV2Watch, runner interface{}, platform, project string, isDevfileTest bool) error {
	var folderToCheck, podName string
	cliRunner := runner.(helper.CliRunner)
	switch platform {
	case "kube":
		if isDevfileTest {
			folderToCheck = "/projects"
			if odoV2Watch.FolderToCheck != "" {
				folderToCheck = odoV2Watch.FolderToCheck
			}
			cliRunner := runner.(helper.CliRunner)
			podName = cliRunner.GetRunningPodNameByComponent(odoV2Watch.CmpName, project)

		} else {
			ocRunner := runner.(helper.OcRunner)
			podName = ocRunner.GetRunningPodNameOfComp(odoV1Watch.SrcType+"-app", project)
			envs := ocRunner.GetEnvs(odoV1Watch.SrcType+"-app", odoV1Watch.AppName, project)
			dir := envs["ODO_S2I_SRC_BIN_PATH"]
			folderToCheck = filepath.ToSlash(filepath.Join(dir, "src"))
		}
	default:
		return fmt.Errorf("Platform %s is not supported", platform)
	}

	// helper.MatchAllInOutput(stdOut, []string{"a.txt", ".abc"})
	// helper.DontMatchAllInOutput(stdOut, []string{"abcd"})

	cliRunner.WaitForRunnerCmdOut([]string{"exec", podName, "--namespace", project,
		"--", "ls", "-lai", folderToCheck}, 5, true, func(output string) bool {
		return !(strings.Contains(output, "abcd")) && (strings.Contains(output, "a.txt")) && (strings.Contains(output, ".abc"))
	})

	return nil
}

// ExecCommand executes odo exec with a command
func ExecCommand(context, cmpName string) {
	args := []string{"create", "nodejs", cmpName, "--context", context}
	helper.Cmd("odo", args...).ShouldPass()

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

	args = []string{"push", "--context", context}
	helper.Cmd("odo", args...).ShouldPass()

	args = []string{"exec", "--context", context}
	args = append(args, []string{"--", "touch", "/projects/blah.js"}...)
	helper.Cmd("odo", args...).ShouldPass()
}

// ExecCommandWithoutComponentAndDevfileFlag executes odo exec without a component and with a devfile flag
func ExecCommandWithoutComponentAndDevfileFlag(context, cmpName string) {
	args := []string{"create", "nodejs", cmpName, "--context", context}
	helper.Cmd("odo", args...).ShouldPass()

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

	args = []string{"exec", "--context", context}
	args = append(args, []string{"--", "touch", "/projects/blah.js"}...)
	helper.Cmd("odo", args...).ShouldFail()

	args = []string{"exec", "--context", context, "--devfile", "invalid.yaml"}
	args = append(args, []string{"--", "touch", "/projects/blah.js"}...)
	helper.Cmd("odo", args...).ShouldFail()
}

//ExecWithoutCommand executes odo exec with no user command and fails
func ExecWithoutCommand(context, cmpName string) {
	args := []string{"create", "nodejs", cmpName, "--context", context}
	helper.Cmd("odo", args...).ShouldPass()

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

	args = []string{"push", "--context", context}
	helper.Cmd("odo", args...).ShouldPass()

	args = []string{"exec", "--context", context}
	args = append(args, "--")
	output := helper.Cmd("odo", args...).ShouldFail().Err()

	Expect(output).To(ContainSubstring("no command was given"))

}

//ExecWithInvalidCommand executes odo exec with a invalid command
func ExecWithInvalidCommand(context, cmpName, pushTarget string) {
	args := []string{"create", "nodejs", cmpName, "--context", context}
	helper.Cmd("odo", args...).ShouldPass()

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

	args = []string{"push", "--context", context}
	helper.Cmd("odo", args...).ShouldPass()

	args = []string{"exec", "--context", context}
	args = append(args, "--", "invalidCommand")
	output := helper.Cmd("odo", args...).ShouldFail().Err()

	Expect(output).To(ContainSubstring("executable file not found in $PATH"))
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
// in both S2I Component list and Devfile Component list
func VerifyCatalogListComponent(output string, cmpName []string) error {
	var data map[string]interface{}
	listItems := []string{"devfileItems"}

	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return err
	}

	if os.Getenv("KUBERNETES") != "true" {
		listItems = append(listItems, "s2iItems")
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
