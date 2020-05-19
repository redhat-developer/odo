package docker

import (
	"path/filepath"

	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo docker devfile push command tests", func() {
	var cmpName string
	var sourcePath = "/projects/nodejs-web-app"

	var globals helper.Globals

	dockerClient := helper.NewDockerRunner("docker")

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		globals = helper.CommonBeforeEachDocker()
		helper.Chdir(globals.Context)

		cmpName = helper.RandString(6)

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

		helper.CommonAfterEeachDocker(globals)

	})

	Context("Verify devfile push works", func() {

		It("Check that odo push works with a devfile", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", globals.Context, cmpName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), globals.Context)

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml")
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// update devfile and push again
			helper.ReplaceString("devfile.yaml", "name: FOO", "name: BAR")
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml")
		})

		It("Check that odo push works with a devfile that has multiple containers", func() {
			// Springboot devfile references multiple containers
			helper.CmdShouldPass("odo", "create", "java-spring-boot", "--context", globals.Context, cmpName)

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml")
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// update devfile and push again
			helper.ReplaceString("devfile.yaml", "name: FOO", "name: BAR")
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml")
		})

		It("Check that odo push works with a devfile that has volumes defined", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", globals.Context, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), globals.Context)
			helper.RenameFile("devfile.yaml", "devfile-old.yaml")
			helper.RenameFile("devfile-with-volumes.yaml", "devfile.yaml")

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml")
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// Verify the volumes got created successfully (and 3 volumes exist: one source and two defined in devfile)
			label := "component=" + cmpName
			volumes := dockerClient.GetVolumesByLabel(label)
			Expect(len(volumes)).To(Equal(3))
		})

		It("Check that odo push mounts the docker volumes in the container", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", globals.Context, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), globals.Context)
			helper.RenameFile("devfile.yaml", "devfile-old.yaml")
			helper.RenameFile("devfile-with-volumes.yaml", "devfile.yaml")

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml")
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// Retrieve the volume from one of the aliases in the devfile
			volumes := dockerClient.GetVolumesByCompStorageName(cmpName, "myvol")
			Expect(len(volumes)).To(Equal(1))
			vol := volumes[0]

			// Verify the volume is mounted
			volMounted := dockerClient.IsVolumeMountedInContainer(vol, cmpName, "runtime")
			Expect(volMounted).To(Equal(true))
		})

		It("should not build when no changes are detected in the directory and build when a file change is detected", func() {
			utils.ExecPushToTestFileChanges(globals.Context, cmpName, "")
		})

		It("should be able to create a file, push, delete, then push again propagating the deletions", func() {
			newFilePath := filepath.Join(globals.Context, "foobar.txt")
			newDirPath := filepath.Join(globals.Context, "testdir")
			utils.ExecPushWithNewFileAndDir(globals.Context, cmpName, "", newFilePath, newDirPath)

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			containers := dockerClient.GetRunningContainersByCompAlias(cmpName, "runtime")
			Expect(len(containers)).To(Equal(1))

			stdOut := dockerClient.ExecContainer(containers[0], "ls -la "+sourcePath)
			Expect(stdOut).To(ContainSubstring(("foobar.txt")))
			Expect(stdOut).To(ContainSubstring(("testdir")))

			// Now we delete the file and dir and push
			helper.DeleteDir(newFilePath)
			helper.DeleteDir(newDirPath)
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml")

			// Then check to see if it's truly been deleted
			stdOut = dockerClient.ExecContainer(containers[0], "ls -la "+sourcePath)
			Expect(stdOut).To(Not(ContainSubstring(("foobar.txt"))))
			Expect(stdOut).To(Not(ContainSubstring(("testdir"))))
		})

		It("should build when no changes are detected in the directory and force flag is enabled", func() {
			utils.ExecPushWithForceFlag(globals.Context, cmpName, "")
		})

		It("should execute the default devbuild and devrun commands if present", func() {
			utils.ExecDefaultDevfileCommands(globals.Context, cmpName, "")

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			containers := dockerClient.GetRunningContainersByCompAlias(cmpName, "runtime")
			Expect(len(containers)).To(Equal(1))

			stdOut := dockerClient.ExecContainer(containers[0], "ps -ef")
			Expect(stdOut).To(ContainSubstring(("/myproject/app.jar")))
		})

		It("should execute the optional devinit, and devrun commands if present", func() {
			helper.CmdShouldPass("odo", "create", "java-spring-boot", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), globals.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile-init.yaml"), filepath.Join(globals.Context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml")
			Expect(output).To(ContainSubstring("Executing devinit command \"echo hello"))
			Expect(output).To(ContainSubstring("Executing devbuild command \"/artifacts/bin/build-container-full.sh\""))
			Expect(output).To(ContainSubstring("Executing devrun command \"/artifacts/bin/start-server.sh\""))

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			containers := dockerClient.GetRunningContainersByCompAlias(cmpName, "runtime")
			Expect(len(containers)).To(Equal(1))

			stdOut := dockerClient.ExecContainer(containers[0], "ps -ef")
			Expect(stdOut).To(ContainSubstring(("/myproject/app.jar")))
		})

		It("should execute devinit and devrun commands if present", func() {
			helper.CmdShouldPass("odo", "create", "java-spring-boot", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), globals.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile-init-without-build.yaml"), filepath.Join(globals.Context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml")
			Expect(output).To(ContainSubstring("Executing devinit command \"echo hello"))
			Expect(output).To(ContainSubstring("Executing devrun command \"/artifacts/bin/start-server.sh\""))

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			containers := dockerClient.GetRunningContainersByCompAlias(cmpName, "runtime")
			Expect(len(containers)).To(Equal(1))

			stdOut := dockerClient.ExecContainer(containers[0], "ls /data")
			Expect(stdOut).To(ContainSubstring(("afile.txt")))
		})

		It("should be able to handle a missing devbuild command", func() {
			utils.ExecWithMissingBuildCommand(globals.Context, cmpName, "")
		})

		It("should error out on a missing devrun command", func() {
			utils.ExecWithMissingRunCommand(globals.Context, cmpName, "")
		})

		It("should be able to push using the custom commands", func() {
			utils.ExecWithCustomCommand(globals.Context, cmpName, "")
		})

		It("should error out on a wrong custom commands", func() {
			utils.ExecWithWrongCustomCommand(globals.Context, cmpName, "")
		})

	})

})
