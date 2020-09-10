package devfile

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/odo/pkg/util"
	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile push command tests", func() {
	var namespace, context, cmpName, currentWorkingDirectory, originalKubeconfig string
	var sourcePath = "/projects"

	// Using program commmand according to cliRunner in devfile
	cliRunner := helper.GetCliRunner()

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		originalKubeconfig = os.Getenv("KUBECONFIG")
		helper.LocalKubeconfigSet(context)
		namespace = cliRunner.CreateRandNamespaceProject()
		currentWorkingDirectory = helper.Getwd()
		cmpName = helper.RandString(6)
		helper.Chdir(context)
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		cliRunner.DeleteNamespaceProject(namespace)
		helper.Chdir(currentWorkingDirectory)
		err := os.Setenv("KUBECONFIG", originalKubeconfig)
		Expect(err).NotTo(HaveOccurred())
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("Pushing devfile without an .odo folder", func() {

		It("should be able to push based on metadata.name in devfile WITH a dash in the name", func() {
			// This is the name that's contained within `devfile-with-metadataname-foobar.yaml`
			name := "foobar"
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile-with-metadataname-foobar.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--namespace", namespace)
			Expect(output).To(ContainSubstring("Executing devfile commands for component " + name))
		})

		It("should be able to push based on name passed", func() {
			name := "springboot"
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--namespace", namespace, name)
			Expect(output).To(ContainSubstring("Executing devfile commands for component " + name))
		})

		It("should error out on devfile flag", func() {
			helper.CmdShouldFail("odo", "push", "--namespace", namespace, "--devfile", "invalid.yaml")
		})

	})

	Context("Verify devfile push works", func() {

		It("should have no errors when no endpoints within the devfile, should create a service when devfile has endpoints", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.RenameFile("devfile.yaml", "devfile-old.yaml")
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-no-endpoints.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push", "--project", namespace)
			output := cliRunner.GetServices(namespace)
			Expect(output).NotTo(ContainSubstring(cmpName))

			helper.RenameFile("devfile-old.yaml", "devfile.yaml")
			output = helper.CmdShouldPass("odo", "push", "--project", namespace)

			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
			output = cliRunner.GetServices(namespace)
			Expect(output).To(ContainSubstring(cmpName))
		})

		It("checks that odo push works with a devfile", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--project", namespace)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// update devfile and push again
			helper.ReplaceString("devfile.yaml", "name: FOO", "name: BAR")
			helper.CmdShouldPass("odo", "push", "--project", namespace)
		})

		It("checks that odo push works with a devfile with sourcemapping set", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileSourceMapping.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--project", namespace)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// Verify source code was synced to /test instead of /projects
			var statErr error
			podName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)
			cliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime",
				namespace,
				[]string{"stat", "/test/server.js"},
				func(cmdOp string, err error) bool {
					statErr = err
					return err == nil
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
		})

		It("checks that odo push works with a devfile with composite commands", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeCommands.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--context", context)
			Expect(output).To(ContainSubstring("Executing mkdir command"))

			// Verify the command executed successfully
			var statErr error
			podName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)
			cliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime",
				namespace,
				[]string{"stat", "/projects/testfolder"},
				func(cmdOp string, err error) bool {
					statErr = err
					return err == nil
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
		})

		It("checks that odo push works with a devfile with parallel composite commands", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeCommandsParallel.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--build-command", "buildAndMkdir", "--context", context)
			Expect(output).To(ContainSubstring("Executing mkdir command"))

			// Verify the command executed successfully
			var statErr error
			podName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)
			cliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime",
				namespace,
				[]string{"stat", "/projects/testfolder"},
				func(cmdOp string, err error) bool {
					statErr = err
					return err == nil
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
		})

		It("checks that odo push works with a devfile with nested composite commands", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileNestedCompCommands.yaml"), filepath.Join(context, "devfile.yaml"))

			// Verify nested command was executed
			output := helper.CmdShouldPass("odo", "push", "--context", context)
			Expect(output).To(ContainSubstring("Executing mkdir command"))

			// Verify the command executed successfully
			var statErr error
			podName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)
			cliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime",
				namespace,
				[]string{"stat", "/projects/testfolder"},
				func(cmdOp string, err error) bool {
					statErr = err
					return err == nil
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
		})

		It("should throw a validation error for composite run commands", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeRun.yaml"), filepath.Join(context, "devfile.yaml"))

			// Verify odo push failed
			output := helper.CmdShouldFail("odo", "push", "--context", context)
			Expect(output).To(ContainSubstring("composite commands of run Kind are not supported currently"))
		})

		It("should throw a validation error for composite command referencing non-existent commands", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeNonExistent.yaml"), filepath.Join(context, "devfile.yaml"))

			// Verify odo push failed
			output := helper.CmdShouldFail("odo", "push", "--context", context)
			Expect(output).To(ContainSubstring("does not exist in the devfile"))
		})

		It("should throw a validation error for composite command indirectly referencing itself", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileIndirectNesting.yaml"), filepath.Join(context, "devfile.yaml"))

			// Verify odo push failed
			output := helper.CmdShouldFail("odo", "push", "--context", context)
			Expect(output).To(ContainSubstring("cannot indirectly reference itself"))
		})

		It("should throw a validation error for composite command that has invalid exec subcommand", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeInvalidComponent.yaml"), filepath.Join(context, "devfile.yaml"))

			// Verify odo push failed
			output := helper.CmdShouldFail("odo", "push", "--context", context)
			Expect(output).To(ContainSubstring("references an invalid command"))
		})

		It("checks that odo push works outside of the context directory", func() {
			helper.Chdir(currentWorkingDirectory)

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, "--context", context, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--context", context)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
		})

		It("should not build when no changes are detected in the directory and build when a file change is detected", func() {
			utils.ExecPushToTestFileChanges(context, cmpName, namespace)
		})

		It("checks that odo push with -o json displays machine readable JSON event output", func() {

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "-o", "json", "--project", namespace)
			utils.AnalyzePushConsoleOutput(output)

			// update devfile and push again
			helper.ReplaceString("devfile.yaml", "name: FOO", "name: BAR")
			output = helper.CmdShouldPass("odo", "push", "-o", "json", "--project", namespace)
			utils.AnalyzePushConsoleOutput(output)

		})

		It("should be able to create a file, push, delete, then push again propagating the deletions", func() {
			newFilePath := filepath.Join(context, "foobar.txt")
			newDirPath := filepath.Join(context, "testdir")
			utils.ExecPushWithNewFileAndDir(context, cmpName, namespace, newFilePath, newDirPath)

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)

			stdOut := cliRunner.ExecListDir(podName, namespace, sourcePath)
			helper.MatchAllInOutput(stdOut, []string{"foobar.txt", "testdir"})

			// Now we delete the file and dir and push
			helper.DeleteDir(newFilePath)
			helper.DeleteDir(newDirPath)
			helper.CmdShouldPass("odo", "push", "--project", namespace, "-v4")

			// Then check to see if it's truly been deleted
			stdOut = cliRunner.ExecListDir(podName, namespace, sourcePath)
			helper.DontMatchAllInOutput(stdOut, []string{"foobar.txt", "testdir"})
		})

		It("should delete the files from the container if its removed locally", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push", "--project", namespace)

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)

			var statErr error
			cliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"",
				namespace,
				[]string{"stat", "/projects/server.js"},
				func(cmdOp string, err error) bool {
					statErr = err
					return err == nil
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
			Expect(os.Remove(filepath.Join(context, "server.js"))).NotTo(HaveOccurred())
			helper.CmdShouldPass("odo", "push", "--project", namespace)

			cliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"",
				namespace,
				[]string{"stat", "/projects/server.js"},
				func(cmdOp string, err error) bool {
					statErr = err
					return err == nil
				},
			)
			Expect(statErr).To(HaveOccurred())
			Expect(statErr.Error()).To(ContainSubstring("cannot stat '/projects/server.js': No such file or directory"))
		})

		It("should build when no changes are detected in the directory and force flag is enabled", func() {
			utils.ExecPushWithForceFlag(context, cmpName, namespace)
		})

		It("should execute the default build and run command groups if present", func() {
			utils.ExecDefaultDevfileCommands(context, cmpName, namespace)

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)

			var statErr error
			var cmdOutput string
			cliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime",
				namespace,
				[]string{"ps", "-ef"},
				func(cmdOp string, err error) bool {
					cmdOutput = cmdOp
					statErr = err
					return err == nil
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
			Expect(cmdOutput).To(ContainSubstring("/myproject/app.jar"))
		})

		It("should execute PreStart commands if present during pod startup", func() {
			expectedInitContainers := []string{"tools-myprestart-1", "tools-myprestart-2", "runtime-secondprestart-3"}

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-valid-events.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--namespace", namespace)
			helper.MatchAllInOutput(output, []string{"PreStart commands have been added to the component"})

			firstPushPodName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)

			firstPushInitContainers := cliRunner.GetPodInitContainers(cmpName, namespace)
			// 3 preStart events + 1 supervisord init containers
			Expect(len(firstPushInitContainers)).To(Equal(4))
			helper.MatchAllInOutput(strings.Join(firstPushInitContainers, ","), expectedInitContainers)

			// Need to force so build and run get triggered again with the component already created.
			output = helper.CmdShouldPass("odo", "push", "--namespace", namespace, "-f")
			helper.MatchAllInOutput(output, []string{"PreStart commands have been added to the component"})

			secondPushPodName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)

			secondPushInitContainers := cliRunner.GetPodInitContainers(cmpName, namespace)
			// 3 preStart events + 1 supervisord init containers
			Expect(len(secondPushInitContainers)).To(Equal(4))
			helper.MatchAllInOutput(strings.Join(secondPushInitContainers, ","), expectedInitContainers)

			Expect(firstPushPodName).To(Equal(secondPushPodName))
			Expect(firstPushInitContainers).To(Equal(secondPushInitContainers))

			var statErr error
			cliRunner.CheckCmdOpInRemoteDevfilePod(
				firstPushPodName,
				"runtime",
				namespace,
				[]string{"cat", "/projects/test.txt"},
				func(cmdOp string, err error) bool {
					if err != nil {
						statErr = err
					} else if cmdOp == "" {
						statErr = fmt.Errorf("prestart event action error, expected: hello test2\nhello test2\nhello test\n, got empty string")
					} else {
						fileContents := strings.Split(cmdOp, "\n")
						if len(fileContents)-1 != 3 {
							statErr = fmt.Errorf("prestart event action count error, expected: 3 strings, got %d strings: %s", len(fileContents), strings.Join(fileContents, ","))
						} else if cmdOp != "hello test2\nhello test2\nhello test\n" {
							statErr = fmt.Errorf("prestart event action error, expected: hello test2\nhello test2\nhello test\n, got: %s", cmdOp)
						}
					}

					return true
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
		})

		It("should execute PostStart commands if present and not execute when component already exists", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-valid-events.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--namespace", namespace)
			helper.MatchAllInOutput(output, []string{"Executing mypoststart command \"echo I am a PostStart\"", "Executing secondpoststart command \"echo I am also a PostStart\""})

			// Need to force so build and run get triggered again with the component already created.
			output = helper.CmdShouldPass("odo", "push", "--namespace", namespace, "-f")
			helper.DontMatchAllInOutput(output, []string{"Executing mypoststart command \"echo I am a PostStart\"", "Executing secondpoststart command \"echo I am also a PostStart\""})
			helper.MatchAllInOutput(output, []string{
				"Executing devbuild command",
				"Executing devrun command",
			})
		})

		It("should err out on an event not mentioned in the devfile commands", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-valid-events.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.ReplaceString("devfile.yaml", "secondpoststart", "secondpoststart12345")

			output := helper.CmdShouldFail("odo", "push", "--namespace", namespace)
			helper.MatchAllInOutput(output, []string{"does not map to a valid devfile command"})
		})

		It("should err out on an event command not mapping to a devfile container component", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-valid-events.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.ReplaceString("devfile.yaml", "secondpoststart", "wrongPostStart")

			output := helper.CmdShouldFail("odo", "push", "--namespace", namespace)
			helper.MatchAllInOutput(output, []string{"the command does not map to a supported component"})
		})

		It("should err out on an event composite command mentioning an invalid child command", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-valid-events.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.ReplaceString("devfile.yaml", "secondpoststart", "myWrongCompCmd")

			output := helper.CmdShouldFail("odo", "push", "--namespace", namespace)
			helper.MatchAllInOutput(output, []string{"does not exist in the devfile"})
		})

		It("should be able to handle a missing build command group", func() {
			utils.ExecWithMissingBuildCommand(context, cmpName, namespace)
		})

		It("should error out on a missing run command group", func() {
			utils.ExecWithMissingRunCommand(context, cmpName, namespace)
		})

		It("should be able to push using the custom commands", func() {
			utils.ExecWithCustomCommand(context, cmpName, namespace)
		})

		It("should error out on a wrong custom commands", func() {
			utils.ExecWithWrongCustomCommand(context, cmpName, namespace)
		})

		It("should error out on multiple or no default commands", func() {
			utils.ExecWithMultipleOrNoDefaults(context, cmpName, namespace)
		})

		It("should execute commands with flags if there are more than one default command", func() {
			utils.ExecMultipleDefaultsWithFlags(context, cmpName, namespace)
		})

		It("should execute commands with flags if the command has no group kind", func() {
			utils.ExecCommandWithoutGroupUsingFlags(context, cmpName, namespace)
		})

		It("should error out if the devfile has an invalid command group", func() {
			utils.ExecWithInvalidCommandGroup(context, cmpName, namespace)
		})

		It("should restart the application if it is not hot reload capable", func() {
			utils.ExecWithHotReload(context, cmpName, namespace, false)
		})

		It("should not restart the application if it is hot reload capable", func() {
			utils.ExecWithHotReload(context, cmpName, namespace, true)
		})

		It("should restart the application if run mode is changed, regardless of hotReloadCapable value", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-hotReload.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push", "--namespace", namespace)

			helper.CmdShouldPass("odo", "push", "--debug", "--namespace", namespace)

			logs := helper.CmdShouldPass("odo", "log")

			helper.MatchAllInOutput(logs, []string{
				"\"stop the program\" program=debugrun",
				"\"stop the program\" program=devrun",
			})

		})

		It("should run odo push successfully after odo push --debug", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--debug", "--namespace", namespace)
			helper.MatchAllInOutput(output, []string{
				"Executing devbuild command",
				"Executing debugrun command",
			})

			output = helper.CmdShouldPass("odo", "push", "--namespace", namespace)
			helper.MatchAllInOutput(output, []string{
				"Executing devbuild command",
				"Executing devrun command",
			})

		})

		It("should create pvc and reuse if it shares the same devfile volume name", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volumes.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--namespace", namespace)
			helper.MatchAllInOutput(output, []string{
				"Executing devbuild command",
				"Executing devrun command",
			})

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)

			var statErr error
			var cmdOutput string

			cliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime2",
				namespace,
				[]string{"cat", "/data/myfile.log"},
				func(cmdOp string, err error) bool {
					cmdOutput = cmdOp
					statErr = err
					return err == nil
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
			Expect(cmdOutput).To(ContainSubstring("hello"))

			cliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime2",
				namespace,
				[]string{"stat", "/data2"},
				func(cmdOp string, err error) bool {
					statErr = err
					return err == nil
				},
			)
			Expect(statErr).ToNot(HaveOccurred())

			volumesMatched := false

			// check the volume name and mount paths for the containers
			volNamesAndPaths := cliRunner.GetVolumeMountNamesandPathsFromContainer(cmpName, "runtime", namespace)
			volNamesAndPathsArr := strings.Fields(volNamesAndPaths)
			for _, volNamesAndPath := range volNamesAndPathsArr {
				volNamesAndPathArr := strings.Split(volNamesAndPath, ":")

				if strings.Contains(volNamesAndPathArr[0], "myvol") && volNamesAndPathArr[1] == "/data" {
					volumesMatched = true
				}
			}
			Expect(volumesMatched).To(Equal(true))
		})

		It("Ensure that push -f correctly removes local deleted files from the remote target sync folder", func() {

			// 1) Push a generic Java project
			helper.CmdShouldPass("odo", "create", "java-springboot", "--project", namespace, cmpName)
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), context)

			output := helper.CmdShouldPass("odo", "push", "--namespace", namespace)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// 2) Rename the pom.xml, which should cause the build to fail if sync is working as expected
			err := os.Rename(filepath.Join(context, "pom.xml"), filepath.Join(context, "pom.xml.renamed"))
			Expect(err).NotTo(HaveOccurred())

			// 3) Ensure that the build fails due to missing 'pom.xml', which ensures that the sync operation
			// correctly renamed pom.xml to pom.xml.renamed.
			session := helper.CmdRunner("odo", "push", "-v", "5", "-f", "--namespace", namespace)
			helper.WaitForOutputToContain("Non-readable POM", 180, 10, session)

		})

	})

	Context("Verify files are correctly synced", func() {

		// Tests https://github.com/openshift/odo/issues/3838
		ensureFilesSyncedTest := func(namespace string, shouldForcePush bool) {
			helper.CmdShouldPass("odo", "create", "java-springboot", "--project", namespace, cmpName)
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), context)

			fmt.Fprintf(GinkgoWriter, "Testing with force push %v", shouldForcePush)

			// 1) Push a standard spring boot project
			output := helper.CmdShouldPass("odo", "push", "--namespace", namespace)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// 2) Update the devfile.yaml, causing push to redeploy the component
			helper.ReplaceString("devfile.yaml", "memoryLimit: 768Mi", "memoryLimit: 769Mi")
			commands := []string{"push", "-v", "4", "--namespace", namespace}
			if shouldForcePush {
				// Test both w/ and w/o '-f'
				commands = append(commands, "-f")
			}

			// 3) Ensure the build passes, indicating that all files were correctly synced to the new pod
			output = helper.CmdShouldPass("odo", commands...)
			Expect(output).To(ContainSubstring("BUILD SUCCESS"))

			// 4) Acquire files from remote container, filtering out target/* and .*
			podName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)
			output = cliRunner.Exec(podName, namespace, "find", sourcePath)
			remoteFiles := []string{}
			outputArr := strings.Split(output, "\n")
			for _, line := range outputArr {

				if !strings.HasPrefix(line, sourcePath+"/") {
					continue
				}

				newLine, err := filepath.Rel(sourcePath, line)
				Expect(err).ToNot(HaveOccurred())

				newLine = filepath.ToSlash(newLine)
				if strings.HasPrefix(newLine, "target/") || newLine == "target" || strings.HasPrefix(newLine, ".") {
					continue
				}

				remoteFiles = append(remoteFiles, newLine)
			}

			// 5) Acquire file from local context, filtering out .*
			localFiles := []string{}
			err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				newPath := filepath.ToSlash(path)

				if strings.HasPrefix(newPath, ".") {
					return nil
				}

				localFiles = append(localFiles, newPath)
				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			// 6) Sort and compare the local and remote files; they should match
			sort.Strings(localFiles)
			sort.Strings(remoteFiles)
			Expect(localFiles).To(Equal(remoteFiles))
		}

		It("Should ensure that files are correctly synced on pod redeploy, with force push specified", func() {
			ensureFilesSyncedTest(namespace, true)
		})

		It("Should ensure that files are correctly synced on pod redeploy, without force push specified", func() {
			ensureFilesSyncedTest(namespace, false)
		})

	})

	Context("Verify devfile volume components work", func() {

		It("should error out when duplicate volume components exist", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.RenameFile("devfile.yaml", "devfile-old.yaml")
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volume-components.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.ReplaceString("devfile.yaml", "secondvol", "firstvol")

			output := helper.CmdShouldFail("odo", "push", "--project", namespace)
			Expect(output).To(ContainSubstring("duplicate volume components present in devfile"))
		})

		It("should error out when a wrong volume size is used", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.RenameFile("devfile.yaml", "devfile-old.yaml")
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volume-components.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.ReplaceString("devfile.yaml", "3Gi", "3Garbage")

			output := helper.CmdShouldFail("odo", "push", "--project", namespace)
			Expect(output).To(ContainSubstring("quantities must match the regular expression"))
		})

		It("should error out if a container component has volume mount that does not refer a valid volume component", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.RenameFile("devfile.yaml", "devfile-old.yaml")
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-invalid-volmount.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldFail("odo", "push", "--project", namespace)
			Expect(output).To(ContainSubstring("unable to find volume mount"))
		})

		It("should successfully use the volume components in container components", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.RenameFile("devfile.yaml", "devfile-old.yaml")
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volume-components.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--project", namespace)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// Verify the pvc size for firstvol
			storageSize := cliRunner.GetPVCSize(cmpName, "firstvol", namespace)
			// should be the default size
			Expect(storageSize).To(ContainSubstring("1Gi"))

			// Verify the pvc size for secondvol
			storageSize = cliRunner.GetPVCSize(cmpName, "secondvol", namespace)
			// should be the specified size in the devfile volume component
			Expect(storageSize).To(ContainSubstring("3Gi"))
		})

		It("should throw a validation error for v1 devfiles", func() {
			helper.CmdShouldPass("odo", "create", "java-springboot", "--project", namespace, cmpName)

			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV1", "springboot", "devfile-init.yaml"), filepath.Join(context, "devfile.yaml"))

			// Verify odo push failed
			output := helper.CmdShouldFail("odo", "push", "--context", context)
			Expect(output).To(ContainSubstring("this Devfile version is not supported"))
		})

	})

	Context("when .gitignore file exists", func() {
		It("checks that .odo/env exists in gitignore", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			ignoreFilePath := filepath.Join(context, ".gitignore")

			helper.FileShouldContainSubstring(ignoreFilePath, filepath.Join(".odo", "env"))

		})
	})

	Context("exec commands with environment variables", func() {
		It("Should be able to exec command with single environment variable", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-multiple-defaults.yaml"), filepath.Join(context, "devfile.yaml"))
			output := helper.CmdShouldPass("odo", "push", "--build-command", "firstbuild", "--run-command", "singleenv", "--namespace", namespace, "--context", context)
			Expect(output).To(ContainSubstring("mkdir $ENV1"))

			podName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)
			output = cliRunner.ExecListDir(podName, namespace, sourcePath)
			Expect(output).To(ContainSubstring("test_env_variable"))

		})

		It("Should be able to exec command with multiple environment variables", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-multiple-defaults.yaml"), filepath.Join(context, "devfile.yaml"))
			output := helper.CmdShouldPass("odo", "push", "--build-command", "firstbuild", "--run-command", "multipleenv", "--namespace", namespace, "--context", context)
			Expect(output).To(ContainSubstring("mkdir $ENV1 $ENV2"))

			podName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)
			output = cliRunner.ExecListDir(podName, namespace, sourcePath)
			helper.MatchAllInOutput(output, []string{"test_env_variable1", "test_env_variable2"})

		})

		It("Should be able to exec command with environment variable with spaces", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-multiple-defaults.yaml"), filepath.Join(context, "devfile.yaml"))
			output := helper.CmdShouldPass("odo", "push", "--build-command", "firstbuild", "--run-command", "envwithspace", "--namespace", namespace, "--context", context)
			Expect(output).To(ContainSubstring("mkdir \\\"$ENV1\\\""))

			podName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)
			output = cliRunner.ExecListDir(podName, namespace, sourcePath)
			helper.MatchAllInOutput(output, []string{"env with space"})

		})
	})

	Context("push with listing the devfile component", func() {

		It("checks components in a specific app and all apps", func() {

			// component created in "app" application
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, "--context", context, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
			output := helper.CmdShouldPass("odo", "list", "--context", context)
			Expect(helper.Suffocate(output)).To(ContainSubstring(helper.Suffocate(fmt.Sprintf("%s%s%s%sUnpushed", "app", cmpName, namespace, "nodejs"))))

			output = helper.CmdShouldPass("odo", "push", "--context", context)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// component created in different application
			context2 := helper.CreateNewContext()
			cmpName2 := helper.RandString(6)
			appName := helper.RandString(6)

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, "--app", appName, "--context", context2, cmpName2)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context2)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context2, "devfile.yaml"))

			output = helper.CmdShouldPass("odo", "list", "--context", context2)
			Expect(helper.Suffocate(output)).To(ContainSubstring(helper.Suffocate(fmt.Sprintf("%s%s%s%sUnpushed", appName, cmpName2, namespace, "nodejs"))))
			output2 := helper.CmdShouldPass("odo", "push", "--context", context2)
			Expect(output2).To(ContainSubstring("Changes successfully pushed to component"))

			output = helper.CmdShouldPass("odo", "list", "--project", namespace)
			Expect(output).To(ContainSubstring(cmpName))
			Expect(output).ToNot(ContainSubstring(cmpName2))

			output = helper.CmdShouldPass("odo", "list", "--all-apps", "--project", namespace)

			Expect(output).To(ContainSubstring(cmpName))
			Expect(output).To(ContainSubstring(cmpName2))

			helper.DeleteDir(context2)

		})

		It("checks devfile and s2i components together", func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("Skipping test because s2i image is not supported on Kubernetes cluster")
			}

			// component created in "app" application
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, "--context", context, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "list", "--context", context)
			Expect(helper.Suffocate(output)).To(ContainSubstring(helper.Suffocate(fmt.Sprintf("%s%s%s%sUnpushed", "app", cmpName, namespace, "nodejs"))))

			output = helper.CmdShouldPass("odo", "push", "--context", context)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// component created in different application
			context2 := helper.CreateNewContext()
			cmpName2 := helper.RandString(6)
			appName := helper.RandString(6)
			helper.CopyExample(filepath.Join("source", "nodejs"), context2)
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, "--app", appName, "--context", context2, cmpName2)

			output2 := helper.CmdShouldPass("odo", "push", "--context", context2)
			Expect(output2).To(ContainSubstring("Changes successfully pushed to component"))

			output = helper.CmdShouldPass("odo", "list", "--all-apps", "--project", namespace)

			Expect(output).To(ContainSubstring(cmpName))
			Expect(output).To(ContainSubstring(cmpName2))

			output = helper.CmdShouldPass("odo", "list", "--app", appName, "--project", namespace)
			Expect(output).To(Not(ContainSubstring(cmpName))) // cmpName component hasn't been created under appName
			Expect(output).To(ContainSubstring(cmpName2))

			helper.DeleteDir(context2)
		})

	})

	Context("Handle devfiles with parent", func() {
		var server *http.Server
		var freePort int
		var parentTmpFolder string

		var _ = BeforeSuite(func() {
			// get a free port
			var err error
			freePort, err = util.HTTPGetFreePort()
			Expect(err).NotTo(HaveOccurred())

			// move the parent devfiles to a tmp folder
			parentTmpFolder = helper.CreateNewContext()
			helper.CopyExample(filepath.Join("source", "devfiles", "parentSupport"), parentTmpFolder)
			// update the port in the required devfile with the free port
			helper.ReplaceString(filepath.Join(parentTmpFolder, "devfile-middle-layer.yaml"), "(-1)", strconv.Itoa(freePort))

			// start the server and serve from the tmp folder of the devfiles
			server = helper.HttpFileServer(freePort, parentTmpFolder)

			// wait for the server to be respond with the desired result
			helper.HttpWaitFor("http://localhost:"+strconv.Itoa(freePort), "devfile", 10, 1)
		})

		var _ = AfterSuite(func() {
			helper.DeleteDir(parentTmpFolder)
			err := server.Close()
			Expect(err).To(BeNil())
		})

		It("should handle a devfile with a parent and add a extra command", func() {
			utils.ExecPushToTestParent(context, cmpName, namespace)
			podName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)
			listDir := cliRunner.ExecListDir(podName, namespace, "/project/")
			Expect(listDir).To(ContainSubstring("blah.js"))
		})

		It("should handle a parent and override/append it's envs", func() {
			utils.ExecPushWithParentOverride(context, cmpName, namespace, freePort)

			envMap := cliRunner.GetEnvsDevFileDeployment(cmpName, namespace)

			value, ok := envMap["ODO_TEST_ENV_0"]
			Expect(ok).To(BeTrue())
			Expect(value).To(Equal("ENV_VALUE_0"))

			value, ok = envMap["ODO_TEST_ENV_1"]
			Expect(ok).To(BeTrue())
			Expect(value).To(Equal("ENV_VALUE_1_1"))
		})

		It("should handle a multi layer parent", func() {
			utils.ExecPushWithMultiLayerParent(context, cmpName, namespace, freePort)

			podName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)
			listDir := cliRunner.ExecListDir(podName, namespace, "/project")
			helper.MatchAllInOutput(listDir, []string{"blah.js", "new-blah.js"})

			envMap := cliRunner.GetEnvsDevFileDeployment(cmpName, namespace)

			value, ok := envMap["ODO_TEST_ENV_1"]
			Expect(ok).To(BeTrue())
			Expect(value).To(Equal("ENV_VALUE_1_1"))

			value, ok = envMap["ODO_TEST_ENV_2"]
			Expect(ok).To(BeTrue())
			Expect(value).To(Equal("ENV_VALUE_2"))

			value, ok = envMap["ODO_TEST_ENV_3"]
			Expect(ok).To(BeTrue())
			Expect(value).To(Equal("ENV_VALUE_3"))

		})
	})

})
