package utils

import (
	"encoding/json"
	"fmt"
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
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.CmdShouldPass("odo", args...)
	helper.MatchAllInOutput(output, []string{
		"Executing defaultbuild command \"/artifacts/bin/build-container-full.sh\"",
		"Executing defaultrun command \"/artifacts/bin/start-server.sh\"",
	})
}

// ExecWithMissingBuildCommand executes odo push with a missing build command
func ExecWithMissingBuildCommand(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-without-devbuild.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.CmdShouldPass("odo", args...)
	Expect(output).NotTo(ContainSubstring("Executing devbuild command"))
	Expect(output).To(ContainSubstring("Executing devrun command \"npm install && npm start\""))
}

// ExecWithMissingRunCommand executes odo push with a missing run command
func ExecWithMissingRunCommand(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	// Remove the run commands
	helper.ReplaceString(filepath.Join(projectDirPath, "devfile.yaml"), "kind: run", "kind: debug")

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.CmdShouldFail("odo", args...)
	Expect(output).NotTo(ContainSubstring("Executing devrun command"))
	Expect(output).To(ContainSubstring("the command group of kind \"run\" is not found in the devfile"))
}

// ExecWithCustomCommand executes odo push with a custom command
func ExecWithCustomCommand(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push", "--build-command", "build", "--run-command", "run"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.CmdShouldPass("odo", args...)
	helper.MatchAllInOutput(output, []string{
		"Executing build command \"npm install\"",
		"Executing run command \"npm start\"",
	})
}

// ExecWithWrongCustomCommand executes odo push with a wrong custom command
func ExecWithWrongCustomCommand(projectDirPath, cmpName, namespace string) {
	garbageCommand := "buildgarbage"

	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push", "--build-command", garbageCommand}
	args = useProjectIfAvailable(args, namespace)
	output := helper.CmdShouldFail("odo", args...)
	Expect(output).NotTo(ContainSubstring("Executing buildgarbage command"))
	Expect(output).To(ContainSubstring("the command \"%v\" is not found in the devfile", garbageCommand))
}

// ExecWithMultipleOrNoDefaults executes odo push with multiple or no default commands
func ExecWithMultipleOrNoDefaults(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-multiple-defaults.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.CmdShouldFail("odo", args...)
	helper.MatchAllInOutput(output, []string{
		"there should be exactly one default command for command group build, currently there is more than one default command",
		"there should be exactly one default command for command group run, currently there is no default command",
	})
}

// ExecMultipleDefaultsWithFlags executes odo push with multiple default commands using flags
func ExecMultipleDefaultsWithFlags(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-multiple-defaults.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push", "--build-command", "firstbuild", "--run-command", "secondrun"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.CmdShouldPass("odo", args...)
	helper.MatchAllInOutput(output, []string{
		"Executing firstbuild command \"npm install\"",
		"Executing secondrun command \"npm start\"",
	})
}

// ExecCommandWithoutGroupUsingFlags executes odo push with no command group using flags
func ExecCommandWithoutGroupUsingFlags(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-multiple-defaults.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push", "--build-command", "thirdbuild", "--run-command", "secondrun"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.CmdShouldPass("odo", args...)
	helper.MatchAllInOutput(output, []string{
		"Executing thirdbuild command \"npm install\"",
		"Executing secondrun command \"npm start\"",
	})
}

// ExecWithInvalidCommandGroup executes odo push with an invalid command group
func ExecWithInvalidCommandGroup(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "java-springboot", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	// Remove the run commands
	helper.ReplaceString(filepath.Join(projectDirPath, "devfile.yaml"), "kind: build", "kind: init")

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.CmdShouldFail("odo", args...)
	Expect(output).To(ContainSubstring("must be one of the following: \"build\", \"run\", \"test\", \"debug\""))
}

func ExecPushToTestParent(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-parent.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	args = append(args, "--build-command", "devBuild", "-f")
	output := helper.CmdShouldPass("odo", args...)
	helper.MatchAllInOutput(output, []string{"Executing devbuild command", "touch blah.js"})
}

func ExecPushWithParentOverride(projectDirPath, cmpName, namespace string, freePort int) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "parentSupport", "devfile-with-parent.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	// update the devfile with the free port
	helper.ReplaceString(filepath.Join(projectDirPath, "devfile.yaml"), "(-1)", strconv.Itoa(freePort))

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)
}

func ExecPushWithMultiLayerParent(projectDirPath, cmpName, namespace string, freePort int) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "parentSupport", "devfile-with-multi-layer-parent.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	// update the devfile with the free port
	helper.ReplaceString(filepath.Join(projectDirPath, "devfile.yaml"), "(-1)", strconv.Itoa(freePort))

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	args = []string{"push", "--build-command", "devbuild", "-f"}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	args = []string{"push", "--build-command", "build", "-f"}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)
}

// ExecPushToTestFileChanges executes odo push with and without a file change
func ExecPushToTestFileChanges(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	output := helper.CmdShouldPass("odo", args...)
	Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

	helper.ReplaceString(filepath.Join(projectDirPath, "server.js"), "Hello from Node.js", "UPDATED!")
	output = helper.CmdShouldPass("odo", args...)
	Expect(output).To(ContainSubstring("Syncing files to the component"))
}

// ExecPushWithForceFlag executes odo push with a force flag
func ExecPushWithForceFlag(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	// use the force build flag and push
	args = []string{"push", "-f"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.CmdShouldPass("odo", args...)
	Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))
}

// ExecPushWithNewFileAndDir executes odo push after creating a new file and dir
func ExecPushWithNewFileAndDir(projectDirPath, cmpName, namespace, newFilePath, newDirPath string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

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
	helper.CmdShouldPass("odo", args...)
}

// ExecWithHotReload executes odo push with hot reload true
func ExecWithHotReload(projectDirPath, cmpName, namespace string, hotReload bool) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)

	if hotReload {
		helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-hotReload.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))
	} else {
		helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))
	}

	args = []string{"push"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.CmdShouldPass("odo", args...)
	Expect(output).To(ContainSubstring("Executing devrun command \"npm start\""))

	args = []string{"push", "-f"}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	args = useProjectIfAvailable([]string{"log"}, namespace)
	logs := helper.CmdShouldPass("odo", args...)
	if hotReload {
		Expect(logs).To(ContainSubstring("Don't start program again, program is already started"))
	} else {
		Expect(logs).To(ContainSubstring("stop the program"))
	}
}

type OdoV1Watch struct {
	SrcType  string
	RouteURL string
	AppName  string
}

type OdoV2Watch struct {
	CmpName            string
	StringsToBeMatched []string
}

// OdoWatch creates files, dir in the context and watches for the changes to be pushed
// Specify OdoV1Watch for odo version 1, OdoV2Watch for odo version 2(devfile)
// platform is either kube or docker
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

			helper.DeleteDir(filepath.Join(context, "abcd"))

			if isDevfileTest {
				helper.ReplaceString(filepath.Join(context, "server.js"), "Hello", "Hello odo")
			} else {
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
					// Verify directory deleted from component pod
					err := validateContainerExecListDir(odoV1Watch, odoV2Watch, runner, platform, project, isDevfileTest)
					Expect(err).To(BeNil())
					return true
				}
			} else {
				curlURL := helper.CmdShouldPass("curl", odoV1Watch.RouteURL)
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
					helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--local-port", freePort)
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

func validateContainerExecListDir(odoV1Watch OdoV1Watch, odoV2Watch OdoV2Watch, runner interface{}, platform, project string, isDevfileTest bool) error {
	var stdOut string

	switch platform {
	case "kube":
		if isDevfileTest {
			cliRunner := runner.(helper.CliRunner)
			podName := cliRunner.GetRunningPodNameByComponent(odoV2Watch.CmpName, project)
			stdOut = cliRunner.ExecListDir(podName, project, "/projects")
		} else {
			ocRunner := runner.(helper.OcRunner)
			podName := ocRunner.GetRunningPodNameOfComp(odoV1Watch.SrcType+"-app", project)
			envs := ocRunner.GetEnvs(odoV1Watch.SrcType+"-app", odoV1Watch.AppName, project)
			dir := envs["ODO_S2I_SRC_BIN_PATH"]
			stdOut = ocRunner.ExecListDir(podName, project, filepath.Join(dir, "src"))
		}
	case "docker":
		dockerRunner := runner.(helper.DockerRunner)
		containers := dockerRunner.GetRunningContainersByCompAlias(odoV2Watch.CmpName, "runtime")
		Expect(len(containers)).To(Equal(1))
		stdOut = dockerRunner.ExecContainer(containers[0], "ls -la /projects")
	default:
		return fmt.Errorf("Platform %s is not supported", platform)
	}

	helper.MatchAllInOutput(stdOut, []string{"a.txt", ".abc"})
	helper.DontMatchAllInOutput(stdOut, []string{"abcd"})

	return nil
}

// ExecCommand executes odo exec with a command
func ExecCommand(context, cmpName string) {
	args := []string{"create", "nodejs", cmpName, "--context", context}
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

	args = []string{"push", "--context", context}
	helper.CmdShouldPass("odo", args...)

	args = []string{"exec", "--context", context}
	args = append(args, []string{"--", "touch", "/projects/blah.js"}...)
	helper.CmdShouldPass("odo", args...)
}

// ExecCommandWithoutComponentAndDevfileFlag executes odo exec without a component and with a devfile flag
func ExecCommandWithoutComponentAndDevfileFlag(context, cmpName string) {
	args := []string{"create", "nodejs", cmpName, "--context", context}
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

	args = []string{"exec", "--context", context}
	args = append(args, []string{"--", "touch", "/projects/blah.js"}...)
	helper.CmdShouldFail("odo", args...)

	args = []string{"exec", "--context", context, "--devfile", "invalid.yaml"}
	args = append(args, []string{"--", "touch", "/projects/blah.js"}...)
	helper.CmdShouldFail("odo", args...)
}

//ExecWithoutCommand executes odo exec with no user command and fails
func ExecWithoutCommand(context, cmpName string) {
	args := []string{"create", "nodejs", cmpName, "--context", context}
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

	args = []string{"push", "--context", context}
	helper.CmdShouldPass("odo", args...)

	args = []string{"exec", "--context", context}
	args = append(args, "--")
	output := helper.CmdShouldFail("odo", args...)

	Expect(output).To(ContainSubstring("no command was given"))

}

//ExecWithInvalidCommand executes odo exec with a invalid command
func ExecWithInvalidCommand(context, cmpName, pushTarget string) {
	args := []string{"create", "nodejs", cmpName, "--context", context}
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

	args = []string{"push", "--context", context}
	helper.CmdShouldPass("odo", args...)

	args = []string{"exec", "--context", context}
	args = append(args, "--", "invalidCommand")
	var output string

	// since exec package for docker returns no error
	// on execution of an invalid command
	switch strings.ToLower(pushTarget) {
	case "kube":
		output = helper.CmdShouldFail("odo", args...)
	case "docker":
		output = helper.CmdShouldPass("odo", args...)
	}

	Expect(output).To(ContainSubstring("executable file not found in $PATH"))
}

// DeleteLocalConfig helps user to delete local config files with flags
func DeleteLocalConfig(args ...string) {
	helper.CmdShouldFail("odo", args...)
	output := helper.CmdShouldPass("odo", append(args, "-af")...)
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
