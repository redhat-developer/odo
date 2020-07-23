package devfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile push command tests", func() {
	var namespace, context, cmpName, currentWorkingDirectory, originalKubeconfig string
	var sourcePath = "/projects/nodejs-starter"

	// Using program commmand according to cliRunner in devfile
	cliRunner := helper.GetCliRunner()

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		// Devfile push requires experimental mode to be set
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")

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
					return true
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
					return true
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
					return true
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
					return true
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
				[]string{"stat", "/projects/nodejs-starter/server.js"},
				func(cmdOp string, err error) bool {
					statErr = err
					return true
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
			Expect(os.Remove(filepath.Join(context, "server.js"))).NotTo(HaveOccurred())
			helper.CmdShouldPass("odo", "push", "--project", namespace)

			cliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"",
				namespace,
				[]string{"stat", "/projects/nodejs-starter/server.js"},
				func(cmdOp string, err error) bool {
					statErr = err
					return true
				},
			)
			Expect(statErr).To(HaveOccurred())
			Expect(statErr.Error()).To(ContainSubstring("cannot stat '/projects/nodejs-starter/server.js': No such file or directory"))
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
					return true
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
			Expect(cmdOutput).To(ContainSubstring("/myproject/app.jar"))
		})

		// v1 devfile test
		It("should execute devinit command if present in v1 devfiles", func() {
			helper.CmdShouldPass("odo", "create", "java-springboot", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV1", "springboot", "devfile-init.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--namespace", namespace)
			helper.MatchAllInOutput(output, []string{
				"Executing devinit command \"echo hello",
				"Executing devbuild command \"/artifacts/bin/build-container-full.sh\"",
				"Executing devrun command \"/artifacts/bin/start-server.sh\"",
			})
		})

		// v1 devfile test
		It("should execute devinit and devrun commands if present in v1 devfiles", func() {
			helper.CmdShouldPass("odo", "create", "java-springboot", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV1", "springboot", "devfile-init-without-build.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--namespace", namespace)
			helper.MatchAllInOutput(output, []string{
				"Executing devinit command \"echo hello",
				"Executing devrun command \"/artifacts/bin/start-server.sh\"",
			})
		})

		// v1 devfile test
		It("should only execute devinit command once if component is already created in v1 devfiles", func() {
			helper.CmdShouldPass("odo", "create", "java-springboot", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV1", "springboot", "devfile-init.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--namespace", namespace)
			helper.MatchAllInOutput(output, []string{
				"Executing devinit command \"echo hello",
				"Executing devbuild command \"/artifacts/bin/build-container-full.sh\"",
				"Executing devrun command \"/artifacts/bin/start-server.sh\"",
			})

			// Need to force so build and run get triggered again with the component already created.
			output = helper.CmdShouldPass("odo", "push", "--namespace", namespace, "-f")
			Expect(output).NotTo(ContainSubstring("Executing devinit command \"echo hello"))
			helper.MatchAllInOutput(output, []string{
				"Executing devbuild command \"/artifacts/bin/build-container-full.sh\"",
				"Executing devrun command \"/artifacts/bin/start-server.sh\"",
			})
		})

		It("should execute PostStart commands if present", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-post-start.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--namespace", namespace)
			helper.MatchAllInOutput(output, []string{"Executing mypoststart command \"echo I am a PostStart\"", "Executing secondpoststart command \"echo I am also a PostStart\""})
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

		It("should not restart the application if restart is false", func() {
			utils.ExecWithRestartAttribute(context, cmpName, namespace)
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
					return true
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
					return true
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

			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "false")
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
			helper.CmdShouldPass("odo", "preference", "set", "--force", "Experimental", "false")
			helper.CopyExample(filepath.Join("source", "nodejs"), context2)
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, "--app", appName, "--context", context2, cmpName2)

			output2 := helper.CmdShouldPass("odo", "push", "--context", context2)
			Expect(output2).To(ContainSubstring("Changes successfully pushed to component"))

			helper.CmdShouldPass("odo", "preference", "set", "--force", "Experimental", "true")

			output = helper.CmdShouldPass("odo", "list", "--all-apps", "--project", namespace)

			Expect(output).To(ContainSubstring(cmpName))
			Expect(output).To(ContainSubstring(cmpName2))

			output = helper.CmdShouldPass("odo", "list", "--app", appName, "--project", namespace)
			Expect(output).To(Not(ContainSubstring(cmpName))) // cmpName component hasn't been created under appName
			Expect(output).To(ContainSubstring(cmpName2))

			helper.DeleteDir(context2)
		})

	})

})
