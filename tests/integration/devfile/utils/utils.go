package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openshift/odo/tests/helper"

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
	args := []string{"create", "java-spring-boot", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push", "--devfile", "devfile.yaml"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.CmdShouldPass("odo", args...)
	Expect(output).To(ContainSubstring("Executing devbuild command \"/artifacts/bin/build-container-full.sh\""))
	Expect(output).To(ContainSubstring("Executing devrun command \"/artifacts/bin/start-server.sh\""))
}

// ExecWithMissingBuildCommand executes odo push with a missing build command
func ExecWithMissingBuildCommand(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-without-devbuild.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push", "--devfile", "devfile.yaml"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.CmdShouldPass("odo", args...)
	Expect(output).NotTo(ContainSubstring("Executing devbuild command"))
	Expect(output).To(ContainSubstring("Executing devrun command \"npm install && nodemon app.js\""))
}

// ExecWithMissingRunCommand executes odo push with a missing run command
func ExecWithMissingRunCommand(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	// Rename the devrun command
	helper.ReplaceString(filepath.Join(projectDirPath, "devfile.yaml"), "devrun", "randomcommand")

	args = []string{"push", "--devfile", "devfile.yaml"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.CmdShouldFail("odo", args...)
	Expect(output).NotTo(ContainSubstring("Executing devrun command"))
	Expect(output).To(ContainSubstring("The command \"devrun\" was not found in the devfile"))
}

// ExecWithCustomCommand executes odo push with a custom command
func ExecWithCustomCommand(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push", "--devfile", "devfile.yaml", "--build-command", "build", "--run-command", "run"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.CmdShouldPass("odo", args...)
	Expect(output).To(ContainSubstring("Executing build command \"npm install\""))
	Expect(output).To(ContainSubstring("Executing run command \"nodemon app.js\""))
}

// ExecWithWrongCustomCommand executes odo push with a wrong custom command
func ExecWithWrongCustomCommand(projectDirPath, cmpName, namespace string) {
	garbageCommand := "buildgarbage"

	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push", "--devfile", "devfile.yaml", "--build-command", garbageCommand}
	args = useProjectIfAvailable(args, namespace)
	output := helper.CmdShouldFail("odo", args...)
	Expect(output).NotTo(ContainSubstring("Executing buildgarbage command"))
	Expect(output).To(ContainSubstring("The command \"%v\" was not found in the devfile", garbageCommand))
}

// ExecPushToTestFileChanges executes odo push with and without a file change
func ExecPushToTestFileChanges(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push", "--devfile", "devfile.yaml"}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	output := helper.CmdShouldPass("odo", args...)
	Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

	helper.ReplaceString(filepath.Join(projectDirPath, "app", "app.js"), "Hello World!", "UPDATED!")
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

	args = []string{"push", "--devfile", "devfile.yaml"}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	// use the force build flag and push
	args = []string{"push", "--devfile", "devfile.yaml", "-f"}
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
	args = []string{"push", "--devfile", "devfile.yaml"}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)
}

// ExecWithRestartAttribute executes odo push with a command attribute restart
func ExecWithRestartAttribute(projectDirPath, cmpName, namespace string) {
	args := []string{"create", "nodejs", cmpName}
	args = useProjectIfAvailable(args, namespace)
	helper.CmdShouldPass("odo", args...)

	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-restart.yaml"), filepath.Join(projectDirPath, "devfile.yaml"))

	args = []string{"push", "--devfile", "devfile.yaml"}
	args = useProjectIfAvailable(args, namespace)
	output := helper.CmdShouldPass("odo", args...)
	Expect(output).To(ContainSubstring("Executing devrun command \"nodemon app.js\""))

	args = []string{"push", "-f", "--devfile", "devfile.yaml"}
	args = useProjectIfAvailable(args, namespace)
	output = helper.CmdShouldPass("odo", args...)
	Expect(output).To(ContainSubstring("if not running"))

}

// OdoWatch creates files, dir in the context and watches for the changes to be pushed
// this was taken from e2e_images_test.go and modified for devfile
func OdoWatch(cmpName, project, context string, stringsToBeMatched []string, runner interface{}, platform string) {

	startSimulationCh := make(chan bool)
	go func() {
		startMsg := <-startSimulationCh
		if startMsg {
			err := os.MkdirAll(filepath.Join(context, ".abc"), 0750)
			if err != nil {
				panic(err)
			}
			err = os.MkdirAll(filepath.Join(context, "abcd"), 0750)
			if err != nil {
				panic(err)
			}
			_, err = os.Create(filepath.Join(context, "a.txt"))
			if err != nil {
				panic(err)
			}

			helper.DeleteDir(filepath.Join(context, "abcd"))

			helper.ReplaceString(filepath.Join(context, "app", "app.js"), "Hello", "Hello odo")
		}
	}()

	success, err := helper.WatchNonRetCmdStdOut(
		("odo watch --context " + context),
		time.Duration(10)*time.Minute,
		func(output string) bool {
			sourcePath := "/projects/nodejs-web-app"

			stringsMatched := true

			for _, stringToBeMatched := range stringsToBeMatched {
				if !strings.Contains(output, stringToBeMatched) {
					stringsMatched = false
				}
			}

			if stringsMatched {
				// Verify delete from component pod
				stdOut := getContainerExecListdir(runner, platform, cmpName, project, sourcePath)
				Expect(stdOut).To(ContainSubstring(("a.txt")))
				Expect(stdOut).To(ContainSubstring((".abc")))
				Expect(stdOut).To(Not(ContainSubstring(("abcd"))))
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

func getContainerExecListdir(runner interface{}, platform, cmpName, project, sourcePath string) string {
	var stdOut string
	// if !reflect.DeepEqual(runner.(helper.OcRunner), helper.OcRunner{}) {

	if platform == "kube" {
		ocRunner := runner.(helper.OcRunner)
		podName := ocRunner.GetRunningPodNameByComponent(cmpName, project)
		stdOut = ocRunner.ExecListDir(podName, project, sourcePath)
	} else if platform == "docker" {
		dockerRunner := runner.(helper.DockerRunner)
		containers := dockerRunner.GetRunningContainersByCompAlias(cmpName, "runtime")
		Expect(len(containers)).To(Equal(1))
		stdOut = dockerRunner.ExecContainer(containers[0], "ls -la "+sourcePath)
	}

	return stdOut
}
