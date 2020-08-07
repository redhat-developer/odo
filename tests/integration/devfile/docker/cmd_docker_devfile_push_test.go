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
	var sourcePath = "/projects"

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
			helper.CmdShouldPass("odo", "create", "java-springboot", "--context", context, cmpName)

			output := helper.CmdShouldPass("odo", "push")
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// update devfile and push again
			helper.ReplaceString("devfile.yaml", "name: FOO", "name: BAR")
			helper.CmdShouldPass("odo", "push")
		})

		It("Check that odo push works with a devfile that has sourcemapping set", func() {
			// Springboot devfile references multiple containers
			helper.CmdShouldPass("odo", "create", "java-springboot", "--context", context, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileSourceMapping.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push")
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// Verify source code was synced to /test
			containers := dockerClient.GetRunningContainersByCompAlias(cmpName, "runtime")
			Expect(len(containers)).To(Equal(1))

			sourceMapping := "/test"
			stdOut := dockerClient.ExecContainer(containers[0], "ls -la "+sourceMapping)
			helper.MatchAllInOutput(stdOut, []string{"server.js"})
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

			output := helper.CmdShouldPass("odo", "push", "-o", "json")
			utils.AnalyzePushConsoleOutput(output)

			// update devfile and push again
			helper.ReplaceString("devfile.yaml", "name: FOO", "name: BAR")
			output = helper.CmdShouldPass("odo", "push", "-o", "json")
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

		It("should execute the default build and run command groups if present", func() {
			utils.ExecDefaultDevfileCommands(context, cmpName, "")

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			containers := dockerClient.GetRunningContainersByCompAlias(cmpName, "runtime")
			Expect(len(containers)).To(Equal(1))

			stdOut := dockerClient.ExecContainer(containers[0], "ps -ef")
			Expect(stdOut).To(ContainSubstring(("/myproject/app.jar")))
		})

		It("should execute PostStart commands if present", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-valid-events.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push")
			helper.MatchAllInOutput(output, []string{"Executing mypoststart command \"echo I am a PostStart\"", "Executing secondpoststart command \"echo I am also a PostStart\""})
		})

		It("should be able to handle a missing build command group", func() {
			utils.ExecWithMissingBuildCommand(context, cmpName, "")
		})

		It("should error out on a missing run command group", func() {
			utils.ExecWithMissingRunCommand(context, cmpName, "")
		})

		It("should be able to push using the custom commands", func() {
			utils.ExecWithCustomCommand(context, cmpName, "")
		})

		It("should error out on a wrong custom commands", func() {
			utils.ExecWithWrongCustomCommand(context, cmpName, "")
		})

		It("should error out on multiple or no default commands", func() {
			utils.ExecWithMultipleOrNoDefaults(context, cmpName, "")
		})

		It("should execute commands with flags if there are more than one default command", func() {
			utils.ExecMultipleDefaultsWithFlags(context, cmpName, "")
		})

		It("should execute commands with flags if the command has no group kind", func() {
			utils.ExecCommandWithoutGroupUsingFlags(context, cmpName, "")
		})

		It("should error out if the devfile has an invalid command group", func() {
			utils.ExecWithInvalidCommandGroup(context, cmpName, "")
		})

	})

	/*
		Disabled test due to issue https://github.com/openshift/odo/issues/3638

		Context("Handle devfiles with parent", func() {
			It("should handle a devfile with a parent and add a extra command", func() {
				utils.ExecPushToTestParent(context, cmpName, "")
				// Check to see if it's been pushed (foobar.txt abd directory testdir)
				containers := dockerClient.GetRunningContainersByCompAlias(cmpName, "runtime")
				Expect(len(containers)).To(Equal(1))

				stdOut := dockerClient.ExecContainer(containers[0], "ls -a /projects/nodejs-starter")
				Expect(stdOut).To(ContainSubstring(("blah.js")))
			})

			It("should handle a parent and override/append it's envs", func() {
				utils.ExecPushWithParentOverride(context, cmpName, "")
				// Check to see if it's been pushed (foobar.txt abd directory testdir)
				containers := dockerClient.GetRunningContainersByCompAlias(cmpName, "appsodyrun")
				Expect(len(containers)).To(Equal(1))

				envMap := dockerClient.GetEnvsDevFileDeployment(containers[0], "printenv")

				value, ok := envMap["MODE2"]
				Expect(ok).To(BeTrue())
				Expect(value).To(Equal("TEST2-override"))

				value, ok = envMap["myprop-3"]
				Expect(ok).To(BeTrue())
				Expect(value).To(Equal("myval-3"))

				value, ok = envMap["myprop2"]
				Expect(ok).To(BeTrue())
				Expect(value).To(Equal("myval2"))
			})

			It("should handle a multi layer parent", func() {
				utils.ExecPushWithMultiLayerParent(context, cmpName, "")
				// Check to see if it's been pushed (foobar.txt abd directory testdir)
				containers := dockerClient.GetRunningContainersByCompAlias(cmpName, "appsodyrun")
				Expect(len(containers)).To(Equal(1))

				stdOut := dockerClient.ExecContainer(containers[0], "ls -a /projects/user-app")
				helper.MatchAllInOutput(stdOut, []string{"blah.js", "new-blah.js"})

				envMap := dockerClient.GetEnvsDevFileDeployment(containers[0], "printenv")

				value, ok := envMap["MODE2"]
				Expect(ok).To(BeTrue())
				Expect(value).To(Equal("TEST2-override"))

				value, ok = envMap["myprop3"]
				Expect(ok).To(BeTrue())
				Expect(value).To(Equal("myval3"))

				value, ok = envMap["myprop2"]
				Expect(ok).To(BeTrue())
				Expect(value).To(Equal("myval2"))

				value, ok = envMap["myprop4"]
				Expect(ok).To(BeTrue())
				Expect(value).To(Equal("myval4"))
			})
		})
	*/
})
