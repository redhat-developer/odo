package docker

import (
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo docker devfile push command tests", func() {
	var context, currentWorkingDirectory, cmpName string
	var sourcePath = "/projects/nodejs-web-app"

	dockerClient := helper.NewDockerRunner("docker")

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
		cmpName = helper.RandString(6)
		helper.Chdir(context)
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		// Local devfile push requires experimental mode to be set and the pushtarget set to docker
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
		helper.CmdShouldPass("odo", "preference", "set", "pushtarget", "docker")
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		// Stop all containers labeled with the component name
		label := "component=" + cmpName
		dockerClient.StopContainers(label)

		dockerClient.RemoveVolumesByComponent(cmpName)

		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("Verify devfile push works", func() {

		It("Check that odo push works with a devfile", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, cmpName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), context)

			output := helper.CmdShouldPass("odo", "push")
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// update devfile and push again
			helper.ReplaceString("devfile.yaml", "name: FOO", "name: BAR")
			helper.CmdShouldPass("odo", "push")
		})

		It("Check that odo push works with a devfile that has multiple containers", func() {
			// Springboot devfile references multiple containers
			helper.CmdShouldPass("odo", "create", "java-spring-boot", "--context", context, cmpName)

			output := helper.CmdShouldPass("odo", "push")
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// update devfile and push again
			helper.ReplaceString("devfile.yaml", "name: FOO", "name: BAR")
			helper.CmdShouldPass("odo", "push")
		})

		It("Check that odo push works with a devfile that has volumes defined", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), context)
			helper.RenameFile("devfile.yaml", "devfile-old.yaml")
			helper.RenameFile("devfile-with-volumes.yaml", "devfile.yaml")

			output := helper.CmdShouldPass("odo", "push")
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// Verify the volumes got created successfully (and 3 volumes exist: one source and two defined in devfile)
			label := "component=" + cmpName
			volumes := dockerClient.GetVolumesByLabel(label)
			Expect(len(volumes)).To(Equal(4))
		})

		It("Check that odo push mounts the docker volumes in the container", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), context)
			helper.RenameFile("devfile.yaml", "devfile-old.yaml")
			helper.RenameFile("devfile-with-volumes.yaml", "devfile.yaml")

			output := helper.CmdShouldPass("odo", "push")
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// Retrieve the volume from one of the aliases in the devfile
			volumes := dockerClient.GetVolumesByCompStorageName(cmpName, "myvol")
			Expect(len(volumes)).To(Equal(1))
			vol := volumes[0]

			// Verify the volume is mounted
			volMounted := dockerClient.IsVolumeMountedInContainer(vol, cmpName, "runtime")
			Expect(volMounted).To(Equal(true))
		})

		It("checks that odo push with -o json displays machine readable JSON event output", func() {

			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, cmpName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), context)

			output := helper.CmdShouldPass("odo", "push", "-o", "json", "--devfile", "devfile.yaml")
			utils.AnalyzePushConsoleOutput(output)

			// update devfile and push again
			helper.ReplaceString("devfile.yaml", "name: FOO", "name: BAR")
			output = helper.CmdShouldPass("odo", "push", "-o", "json", "--devfile", "devfile.yaml")
			utils.AnalyzePushConsoleOutput(output)

		})

		It("should not build when no changes are detected in the directory and build when a file change is detected", func() {
			utils.ExecPushToTestFileChanges(context, cmpName, "")
		})

		It("should be able to create a file, push, delete, then push again propagating the deletions", func() {
			newFilePath := filepath.Join(context, "foobar.txt")
			newDirPath := filepath.Join(context, "testdir")
			utils.ExecPushWithNewFileAndDir(context, cmpName, "", newFilePath, newDirPath)

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			containers := dockerClient.GetRunningContainersByCompAlias(cmpName, "runtime")
			Expect(len(containers)).To(Equal(1))

			stdOut := dockerClient.ExecContainer(containers[0], "ls -la "+sourcePath)
			helper.MatchAllInOutput(stdOut, []string{"foobar.txt", "testdir"})

			// Now we delete the file and dir and push
			helper.DeleteDir(newFilePath)
			helper.DeleteDir(newDirPath)
			helper.CmdShouldPass("odo", "push")

			// Then check to see if it's truly been deleted
			stdOut = dockerClient.ExecContainer(containers[0], "ls -la "+sourcePath)
			helper.DontMatchAllInOutput(stdOut, []string{"foobar.txt", "testdir"})
		})

		It("should build when no changes are detected in the directory and force flag is enabled", func() {
			utils.ExecPushWithForceFlag(context, cmpName, "")
		})

		It("should execute the default devbuild and devrun commands if present", func() {
			utils.ExecDefaultDevfileCommands(context, cmpName, "")

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			containers := dockerClient.GetRunningContainersByCompAlias(cmpName, "runtime")
			Expect(len(containers)).To(Equal(1))

			stdOut := dockerClient.ExecContainer(containers[0], "ps -ef")
			Expect(stdOut).To(ContainSubstring(("/myproject/app.jar")))
		})

		It("should execute the optional devinit, and devrun commands if present", func() {
			helper.CmdShouldPass("odo", "create", "java-spring-boot", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV1", "springboot", "devfile-init.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push")
			helper.MatchAllInOutput(output, []string{
				"Executing devinit command \"echo hello",
				"Executing devbuild command \"/artifacts/bin/build-container-full.sh\"",
				"Executing devrun command \"/artifacts/bin/start-server.sh\"",
			})

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			containers := dockerClient.GetRunningContainersByCompAlias(cmpName, "runtime")
			Expect(len(containers)).To(Equal(1))

			stdOut := dockerClient.ExecContainer(containers[0], "ps -ef")
			Expect(stdOut).To(ContainSubstring(("/myproject/app.jar")))
		})

		It("should execute devinit and devrun commands if present", func() {
			helper.CmdShouldPass("odo", "create", "java-spring-boot", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV1", "springboot", "devfile-init-without-build.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push")
			helper.MatchAllInOutput(output, []string{
				"Executing devinit command \"echo hello",
				"Executing devrun command \"/artifacts/bin/start-server.sh\"",
			})

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			containers := dockerClient.GetRunningContainersByCompAlias(cmpName, "runtime")
			Expect(len(containers)).To(Equal(1))

			stdOut := dockerClient.ExecContainer(containers[0], "ls /data")
			Expect(stdOut).To(ContainSubstring(("afile.txt")))
		})

		It("should be able to handle a missing devbuild command", func() {
			utils.ExecWithMissingBuildCommand(context, cmpName, "")
		})

		It("should error out on a missing devrun command", func() {
			utils.ExecWithMissingRunCommand(context, cmpName, "")
		})

		It("should be able to push using the custom commands", func() {
			utils.ExecWithCustomCommand(context, cmpName, "")
		})

		It("should error out on a wrong custom commands", func() {
			utils.ExecWithWrongCustomCommand(context, cmpName, "")
		})

	})

})
