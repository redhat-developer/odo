package devfile

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile push command tests", func() {
	var cmpName string
	var sourcePath = "/projects/nodejs-web-app"
	var globals helper.Globals

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		globals = helper.CommonBeforeEach()

		cmpName = helper.RandString(6)
		helper.Chdir(globals.Context)
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")

	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEeach(globals)

	})

	Context("Verify devfile push works", func() {

		It("should have no errors when no endpoints within the devfile, should create a service when devfile has endpoints", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", globals.Project, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), globals.Context)
			helper.RenameFile("devfile.yaml", "devfile-old.yaml")
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-no-endpoints.yaml"), filepath.Join(globals.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", globals.Project)
			output := globals.CliRunner.GetServices(globals.Project)
			Expect(output).NotTo(ContainSubstring(cmpName))

			helper.RenameFile("devfile-old.yaml", "devfile.yaml")
			output = helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", globals.Project)

			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
			output = globals.CliRunner.GetServices(globals.Project)
			Expect(output).To(ContainSubstring(cmpName))
		})

		It("checks that odo push works with a devfile", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", globals.Project, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), globals.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(globals.Context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", globals.Project)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// update devfile and push again
			helper.ReplaceString("devfile.yaml", "name: FOO", "name: BAR")
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", globals.Project)
		})

		It("checks that odo push works outside of the context directory", func() {

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", globals.Project, "--context", globals.Context, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), globals.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(globals.Context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--context", globals.Context)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
		})

		It("should not build when no changes are detected in the directory and build when a file change is detected", func() {
			utils.ExecPushToTestFileChanges(globals.Context, cmpName, globals.Project)
		})

		It("should be able to create a file, push, delete, then push again propagating the deletions", func() {
			newFilePath := filepath.Join(globals.Context, "foobar.txt")
			newDirPath := filepath.Join(globals.Context, "testdir")
			utils.ExecPushWithNewFileAndDir(globals.Context, cmpName, globals.Project, newFilePath, newDirPath)

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := globals.CliRunner.GetRunningPodNameByComponent(cmpName, globals.Project)

			stdOut := globals.CliRunner.ExecListDir(podName, globals.Project, sourcePath)
			Expect(stdOut).To(ContainSubstring(("foobar.txt")))
			Expect(stdOut).To(ContainSubstring(("testdir")))

			// Now we delete the file and dir and push
			helper.DeleteDir(newFilePath)
			helper.DeleteDir(newDirPath)
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", globals.Project, "-v4")

			// Then check to see if it's truly been deleted
			stdOut = globals.CliRunner.ExecListDir(podName, globals.Project, sourcePath)
			Expect(stdOut).To(Not(ContainSubstring(("foobar.txt"))))
			Expect(stdOut).To(Not(ContainSubstring(("testdir"))))
		})

		It("should delete the files from the container if its removed locally", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", globals.Project, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), globals.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(globals.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", globals.Project)

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := globals.CliRunner.GetRunningPodNameByComponent(cmpName, globals.Project)

			var statErr error
			globals.CliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"",
				globals.Project,
				[]string{"stat", "/projects/nodejs-web-app/app/app.js"},
				func(cmdOp string, err error) bool {
					statErr = err
					return true
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
			Expect(os.Remove(filepath.Join(globals.Context, "app", "app.js"))).NotTo(HaveOccurred())
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", globals.Project)

			globals.CliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"",
				globals.Project,
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
			utils.ExecPushWithForceFlag(globals.Context, cmpName, globals.Project)
		})

		It("should execute the default devbuild and devrun commands if present", func() {
			utils.ExecDefaultDevfileCommands(globals.Context, cmpName, globals.Project)

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := globals.CliRunner.GetRunningPodNameByComponent(cmpName, globals.Project)

			var statErr error
			var cmdOutput string
			globals.CliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime",
				globals.Project,
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
			helper.CmdShouldPass("odo", "create", "java-spring-boot", "--project", globals.Project, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), globals.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile-init.yaml"), filepath.Join(globals.Context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", globals.Project)
			Expect(output).To(ContainSubstring("Executing devinit command \"echo hello"))
			Expect(output).To(ContainSubstring("Executing devbuild command \"/artifacts/bin/build-container-full.sh\""))
			Expect(output).To(ContainSubstring("Executing devrun command \"/artifacts/bin/start-server.sh\""))
		})

		It("should execute devinit and devrun commands if present", func() {
			helper.CmdShouldPass("odo", "create", "java-spring-boot", "--project", globals.Project, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), globals.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile-init-without-build.yaml"), filepath.Join(globals.Context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", globals.Project)
			Expect(output).To(ContainSubstring("Executing devinit command \"echo hello"))
			Expect(output).To(ContainSubstring("Executing devrun command \"/artifacts/bin/start-server.sh\""))
		})

		It("should only execute devinit command once if component is already created", func() {
			helper.CmdShouldPass("odo", "create", "java-spring-boot", "--project", globals.Project, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), globals.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile-init.yaml"), filepath.Join(globals.Context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", globals.Project)
			Expect(output).To(ContainSubstring("Executing devinit command \"echo hello"))
			Expect(output).To(ContainSubstring("Executing devbuild command \"/artifacts/bin/build-container-full.sh\""))
			Expect(output).To(ContainSubstring("Executing devrun command \"/artifacts/bin/start-server.sh\""))

			// Need to force so build and run get triggered again with the component already created.
			output = helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", globals.Project, "-f")
			Expect(output).NotTo(ContainSubstring("Executing devinit command \"echo hello"))
			Expect(output).To(ContainSubstring("Executing devbuild command \"/artifacts/bin/build-container-full.sh\""))
			Expect(output).To(ContainSubstring("Executing devrun command \"/artifacts/bin/start-server.sh\""))
		})

		It("should be able to handle a missing devinit command", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", globals.Project, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), globals.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-without-devinit.yaml"), filepath.Join(globals.Context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", globals.Project)
			Expect(output).NotTo(ContainSubstring("Executing devinit command"))
			Expect(output).To(ContainSubstring("Executing devbuild command \"npm install\""))
			Expect(output).To(ContainSubstring("Executing devrun command \"nodemon app.js\""))
		})

		It("should be able to handle a missing devbuild command", func() {
			utils.ExecWithMissingBuildCommand(globals.Context, cmpName, globals.Project)
		})

		It("should error out on a missing devrun command", func() {
			utils.ExecWithMissingRunCommand(globals.Context, cmpName, globals.Project)
		})

		It("should be able to push using the custom commands", func() {
			utils.ExecWithCustomCommand(globals.Context, cmpName, globals.Project)
		})

		It("should error out on a wrong custom commands", func() {
			utils.ExecWithWrongCustomCommand(globals.Context, cmpName, globals.Project)
		})

		It("should not restart the application if restart is false", func() {
			utils.ExecWithRestartAttribute(globals.Context, cmpName, globals.Project)
		})

		It("should create pvc and reuse if it shares the same devfile volume name", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", globals.Project, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), globals.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volumes.yaml"), filepath.Join(globals.Context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", globals.Project)
			Expect(output).To(ContainSubstring("Executing devinit command"))
			Expect(output).To(ContainSubstring("Executing devbuild command"))
			Expect(output).To(ContainSubstring("Executing devrun command"))

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := globals.CliRunner.GetRunningPodNameByComponent(cmpName, globals.Project)

			var statErr error
			var cmdOutput string
			globals.CliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime",
				globals.Project,
				[]string{"cat", "/data/myfile-init.log"},
				func(cmdOp string, err error) bool {
					cmdOutput = cmdOp
					statErr = err
					return true
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
			Expect(cmdOutput).To(ContainSubstring("init"))

			globals.CliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime2",
				globals.Project,
				[]string{"cat", "/data/myfile.log"},
				func(cmdOp string, err error) bool {
					cmdOutput = cmdOp
					statErr = err
					return true
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
			Expect(cmdOutput).To(ContainSubstring("hello"))

			globals.CliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime2",
				globals.Project,
				[]string{"stat", "/data2"},
				func(cmdOp string, err error) bool {
					statErr = err
					return true
				},
			)
			Expect(statErr).ToNot(HaveOccurred())

			volumesMatched := false

			// check the volume name and mount paths for the containers
			volNamesAndPaths := globals.CliRunner.GetVolumeMountNamesandPathsFromContainer(cmpName, "runtime", globals.Project)
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
