package docker

import (
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo docker devfile test command tests", func() {
	var context, currentWorkingDirectory, cmpName string

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

	Context("Should show proper errors", func() {

		It("should show error if component is not pushed", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-testgroup.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldFail("odo", "test", "--context", context)

			Expect(output).To(ContainSubstring("component does not exist, a valid component is required to run 'odo test'"))
		})

		It("should show error if no test group is defined", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
			helper.CmdShouldPass("odo", "push", "--context", context)
			output := helper.CmdShouldFail("odo", "test", "--context", context)

			Expect(output).To(ContainSubstring("the command group of kind \"test\" is not found in the devfile"))
		})

		It("should show error if specify non-exist command", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-testgroup.yaml"), filepath.Join(context, "devfile.yaml"))
			helper.CmdShouldPass("odo", "push", "--context", context)
			output := helper.CmdShouldFail("odo", "test", "--test-command", "invalidcmd", "--context", context)

			Expect(output).To(ContainSubstring("not found in the devfile"))
		})

		It("should show error if command from another group", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-testgroup.yaml"), filepath.Join(context, "devfile.yaml"))
			helper.CmdShouldPass("odo", "push", "--context", context)
			output := helper.CmdShouldFail("odo", "test", "--test-command", "devrun", "--context", context)

			Expect(output).To(ContainSubstring("command devrun is of group run in devfile.yaml"))
		})

		It("should show error if devfile has no default test command", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, cmpName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-testgroup.yaml"), filepath.Join(context, "devfile.yaml"))
			helper.ReplaceString("devfile.yaml", "isDefault: true", "")
			helper.CmdShouldPass("odo", "push", "--context", context)
			output := helper.CmdShouldFail("odo", "test", "--context", context)
			Expect(output).To(ContainSubstring("there should be exactly one default command for command group test, currently there is no default command"))
		})

		It("should show error if devfile has multiple default test command", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, cmpName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-multiple-defaults.yaml"), filepath.Join(context, "devfile.yaml"))
			helper.CmdShouldPass("odo", "push", "--build-command", "firstbuild", "--run-command", "secondrun", "--context", context)
			output := helper.CmdShouldFail("odo", "test", "--context", context)
			Expect(output).To(ContainSubstring("there should be exactly one default command for command group test, currently there is more than one default command"))
		})
	})

	/*

		// commented out because of https://github.com/openshift/odo/issues/3685
		Context("Should run test command successfully", func() {
			const sourcePath = "/projects"
			It("Should run test command successfully with only one default specified", func() {
				helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, cmpName)

				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-testgroup.yaml"), filepath.Join(context, "devfile.yaml"))
				helper.CmdShouldPass("odo", "push", "--context", context)

				output := helper.CmdShouldPass("odo", "test", "--context", context)
				helper.MatchAllInOutput(output, []string{"Executing test1 command", "mkdir test1"})

				containers := dockerClient.GetRunningContainersByCompAlias(cmpName, "runtime")
				Expect(len(containers)).To(Equal(1))
				output = dockerClient.ExecContainer(containers[0], "ls -la "+sourcePath)
				Expect(output).To(ContainSubstring("test1"))
			})

			It("Should run test command successfully with test-command specified", func() {
				helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, cmpName)

				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-testgroup.yaml"), filepath.Join(context, "devfile.yaml"))
				helper.CmdShouldPass("odo", "push", "--context", context)

				output := helper.CmdShouldPass("odo", "test", "--test-command", "test2", "--context", context)
				helper.MatchAllInOutput(output, []string{"Executing test2 command", "mkdir test2"})

				containers := dockerClient.GetRunningContainersByCompAlias(cmpName, "runtime")
				Expect(len(containers)).To(Equal(1))
				output = dockerClient.ExecContainer(containers[0], "ls -la "+sourcePath)
				Expect(output).To(ContainSubstring("test2"))
			})

			It("should run test command successfully with test-command specified if devfile has no default test command", func() {
				helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, cmpName)
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-testgroup.yaml"), filepath.Join(context, "devfile.yaml"))
				helper.ReplaceString("devfile.yaml", "isDefault: true", "")
				helper.CmdShouldPass("odo", "push", "--context", context)
				output := helper.CmdShouldPass("odo", "test", "--test-command", "test2", "--context", context)
				helper.MatchAllInOutput(output, []string{"Executing test2 command", "mkdir test2"})

				containers := dockerClient.GetRunningContainersByCompAlias(cmpName, "runtime")
				Expect(len(containers)).To(Equal(1))
				output = dockerClient.ExecContainer(containers[0], "ls -la "+sourcePath)
				Expect(output).To(ContainSubstring("test2"))
			})

			It("should run test command successfully with test-command specified if devfile has multiple default test command", func() {
				helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, cmpName)
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-multiple-defaults.yaml"), filepath.Join(context, "devfile.yaml"))
				helper.CmdShouldPass("odo", "push", "--build-command", "firstbuild", "--run-command", "secondrun", "--context", context)
				output := helper.CmdShouldPass("odo", "test", "--test-command", "test2", "--context", context)
				helper.MatchAllInOutput(output, []string{"Executing test2 command", "mkdir test2"})

				containers := dockerClient.GetRunningContainersByCompAlias(cmpName, "runtime")
				Expect(len(containers)).To(Equal(1))
				output = dockerClient.ExecContainer(containers[0], "ls -la "+sourcePath)
				Expect(output).To(ContainSubstring("test2"))
			})
		})
	*/

})
