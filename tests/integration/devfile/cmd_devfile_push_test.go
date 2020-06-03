package devfile

import (
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
	var sourcePath = "/projects/nodejs-web-app"

	// Using program commmand according to cliRunner in devfile
	cliRunner := helper.GetCliRunner()

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		// kubeConfigFile := helper.CopyKubeConfigFile(helper.GetExistingKubeConfigPath(), filepath.Join(context, "config"))
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

			output := helper.CmdShouldPass("odo", "push", "-o", "json", "--devfile", "devfile.yaml", "--project", namespace)
			utils.AnalyzePushConsoleOutput(output)

			// update devfile and push again
			helper.ReplaceString("devfile.yaml", "name: FOO", "name: BAR")
			output = helper.CmdShouldPass("odo", "push", "-o", "json", "--devfile", "devfile.yaml", "--project", namespace)
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
				[]string{"stat", "/projects/nodejs-web-app/app/app.js"},
				func(cmdOp string, err error) bool {
					statErr = err
					return true
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
			Expect(os.Remove(filepath.Join(context, "app", "app.js"))).NotTo(HaveOccurred())
			helper.CmdShouldPass("odo", "push", "--project", namespace)

			cliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"",
				namespace,
				[]string{"stat", "/projects/nodejs-web-app/app/app.js"},
				func(cmdOp string, err error) bool {
					statErr = err
					return true
				},
			)
			Expect(statErr).To(HaveOccurred())
			Expect(statErr.Error()).To(ContainSubstring("cannot stat '/projects/nodejs-web-app/app/app.js': No such file or directory"))
		})

		It("should build when no changes are detected in the directory and force flag is enabled", func() {
			utils.ExecPushWithForceFlag(context, cmpName, namespace)
		})

		It("should execute the default devbuild and devrun commands if present", func() {
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

		It("should execute devinit command if present", func() {
			helper.CmdShouldPass("odo", "create", "java-spring-boot", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV1", "springboot", "devfile-init.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--namespace", namespace)
			helper.MatchAllInOutput(output, []string{
				"Executing devinit command \"echo hello",
				"Executing devbuild command \"/artifacts/bin/build-container-full.sh\"",
				"Executing devrun command \"/artifacts/bin/start-server.sh\"",
			})
		})

		It("should execute devinit and devrun commands if present", func() {
			helper.CmdShouldPass("odo", "create", "java-spring-boot", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV1", "springboot", "devfile-init-without-build.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--namespace", namespace)
			helper.MatchAllInOutput(output, []string{
				"Executing devinit command \"echo hello",
				"Executing devrun command \"/artifacts/bin/start-server.sh\"",
			})
		})

		It("should only execute devinit command once if component is already created", func() {
			helper.CmdShouldPass("odo", "create", "java-spring-boot", "--project", namespace, cmpName)

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

		It("should be able to handle a missing devinit command", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-without-devinit.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--namespace", namespace)
			Expect(output).NotTo(ContainSubstring("Executing devinit command"))
			helper.MatchAllInOutput(output, []string{
				"Executing devbuild command \"npm install\"",
				"Executing devrun command \"nodemon app.js\"",
			})
		})

		It("should be able to handle a missing devbuild command", func() {
			utils.ExecWithMissingBuildCommand(context, cmpName, namespace)
		})

		It("should error out on a missing devrun command", func() {
			utils.ExecWithMissingRunCommand(context, cmpName, namespace)
		})

		It("should be able to push using the custom commands", func() {
			utils.ExecWithCustomCommand(context, cmpName, namespace)
		})

		It("should error out on a wrong custom commands", func() {
			utils.ExecWithWrongCustomCommand(context, cmpName, namespace)
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

})
