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
	var namespace, context, cmpName, currentWorkingDirectory string
	var sourcePath = "/projects/nodejs-web-app"

	// TODO: all oc commands in all devfile related test should get replaced by kubectl
	// TODO: to goal is not to use "oc"
	oc := helper.NewOcRunner("oc")

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		namespace = helper.CreateRandProject()
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
		cmpName = helper.RandString(6)

		helper.Chdir(context)

		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		// Devfile push requires experimental mode to be set
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(namespace)
		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("Verify devfile push works", func() {

		It("should have no errors when no endpoints within the devfile, should create a service when devfile has endpoints", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.RenameFile("devfile.yaml", "devfile-old.yaml")
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-no-endpoints.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", namespace)
			output := oc.GetServices(namespace)
			Expect(output).NotTo(ContainSubstring(cmpName))

			helper.RenameFile("devfile-old.yaml", "devfile.yaml")
			output = helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", namespace)

			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
			output = oc.GetServices(namespace)
			Expect(output).To(ContainSubstring(cmpName))
		})

		It("checks that odo push works with a devfile", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", namespace)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// update devfile and push again
			helper.ReplaceString("devfile.yaml", "name: FOO", "name: BAR")
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", namespace)
		})

		It("checks that odo push works outside of the context directory", func() {
			helper.Chdir(currentWorkingDirectory)

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, "--context", context, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--context", context)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
		})

		It("should not build when no changes are detected in the directory and build when a file change is detected", func() {
			utils.ExecPushToTestFileChanges(context, cmpName, namespace)
		})

		It("should be able to create a file, push, delete, then push again propagating the deletions", func() {
			newFilePath := filepath.Join(context, "foobar.txt")
			newDirPath := filepath.Join(context, "testdir")
			utils.ExecPushWithNewFileAndDir(context, cmpName, namespace, newFilePath, newDirPath)

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := oc.GetRunningPodNameByComponent(cmpName, namespace)

			stdOut := oc.ExecListDir(podName, namespace, sourcePath)
			Expect(stdOut).To(ContainSubstring(("foobar.txt")))
			Expect(stdOut).To(ContainSubstring(("testdir")))

			// Now we delete the file and dir and push
			helper.DeleteDir(newFilePath)
			helper.DeleteDir(newDirPath)
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", namespace, "-v4")

			// Then check to see if it's truly been deleted
			stdOut = oc.ExecListDir(podName, namespace, sourcePath)
			Expect(stdOut).To(Not(ContainSubstring(("foobar.txt"))))
			Expect(stdOut).To(Not(ContainSubstring(("testdir"))))
		})

		It("should delete the files from the container if its removed locally", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", namespace)

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := oc.GetRunningPodNameByComponent(cmpName, namespace)

			var statErr error
			oc.CheckCmdOpInRemoteDevfilePod(
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
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", namespace)

			oc.CheckCmdOpInRemoteDevfilePod(
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
			podName := oc.GetRunningPodNameByComponent(cmpName, namespace)

			var statErr error
			var cmdOutput string
			oc.CheckCmdOpInRemoteDevfilePod(
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
			helper.CmdShouldPass("git", "clone", "https://github.com/maysunfaisal/springboot.git", projectDirPath)
			helper.Chdir(projectDirPath)

			helper.CmdShouldPass("odo", "create", "java-spring-boot", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot"), projectDirPath)

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile-init.yaml", "--namespace", namespace)
			Expect(output).To(ContainSubstring("Executing devinit command \"echo hello"))
			Expect(output).To(ContainSubstring("Executing devbuild command \"/artifacts/bin/build-container-full.sh\""))
			Expect(output).To(ContainSubstring("Executing devrun command \"/artifacts/bin/start-server.sh\""))
		})

		It("should execute devinit and devrun commands if present", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/maysunfaisal/springboot.git", projectDirPath)
			helper.Chdir(projectDirPath)

			helper.CmdShouldPass("odo", "create", "java-spring-boot", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot"), projectDirPath)

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile-init-without-build.yaml", "--namespace", namespace)
			Expect(output).To(ContainSubstring("Executing devinit command \"echo hello"))
			Expect(output).To(ContainSubstring("Executing devrun command \"/artifacts/bin/start-server.sh\""))
		})

		It("should only execute devinit command once if component is already created", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/maysunfaisal/springboot.git", projectDirPath)
			helper.Chdir(projectDirPath)

			helper.CmdShouldPass("odo", "create", "java-spring-boot", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot"), projectDirPath)

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile-init.yaml", "--namespace", namespace)
			Expect(output).To(ContainSubstring("Executing devinit command \"echo hello"))
			Expect(output).To(ContainSubstring("Executing devbuild command \"/artifacts/bin/build-container-full.sh\""))
			Expect(output).To(ContainSubstring("Executing devrun command \"/artifacts/bin/start-server.sh\""))

			// Need to force so build and run get triggered again with the component already created.
			output = helper.CmdShouldPass("odo", "push", "--devfile", "devfile-init.yaml", "--namespace", namespace, "-f")
			Expect(output).NotTo(ContainSubstring("Executing devinit command \"echo hello"))
			Expect(output).To(ContainSubstring("Executing devbuild command \"/artifacts/bin/build-container-full.sh\""))
			Expect(output).To(ContainSubstring("Executing devrun command \"/artifacts/bin/start-server.sh\""))
		})

		It("should be able to handle a missing devinit command", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/che-samples/web-nodejs-sample.git", projectDirPath)
			helper.Chdir(projectDirPath)

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), projectDirPath)
			helper.RenameFile("devfile.yaml", "devfile-old.yaml")
			helper.RenameFile("devfile-without-devinit.yaml", "devfile.yaml")

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", namespace)
			Expect(output).NotTo(ContainSubstring("Executing devinit command"))
			Expect(output).To(ContainSubstring("Executing devbuild command \"npm install\""))
			Expect(output).To(ContainSubstring("Executing devrun command \"nodemon app.js\""))
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
			utils.ExecWithRestartAttribute(projectDirPath, cmpName, namespace)
		})

		It("should create pvc and reuse if it shares the same devfile volume name", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volumes.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", namespace)
			Expect(output).To(ContainSubstring("Executing devinit command"))
			Expect(output).To(ContainSubstring("Executing devbuild command"))
			Expect(output).To(ContainSubstring("Executing devrun command"))

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := oc.GetRunningPodNameByComponent(cmpName, namespace)

			var statErr error
			var cmdOutput string
			oc.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime",
				namespace,
				[]string{"cat", "/data/myfile-init.log"},
				func(cmdOp string, err error) bool {
					cmdOutput = cmdOp
					statErr = err
					return true
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
			Expect(cmdOutput).To(ContainSubstring("init"))

			oc.CheckCmdOpInRemoteDevfilePod(
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

			oc.CheckCmdOpInRemoteDevfilePod(
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
			volNamesAndPaths := oc.GetVolumeMountNamesandPathsFromContainer(cmpName, "runtime", namespace)
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
