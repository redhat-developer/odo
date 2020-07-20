package docker

import (
	"github.com/openshift/odo/tests/integration/devfile/utils"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo docker devfile exec command tests", func() {
	var context, currentWorkingDirectory, cmpName string
	var sourcePath = "/projects"

	dockerClient := helper.NewDockerRunner("docker")

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
		cmpName = helper.RandString(6)
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		// Local devfile push requires experimental mode to be set and the pushtarget set to docker
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true", "-f")
		helper.CmdShouldPass("odo", "preference", "set", "pushtarget", "docker", "-f")
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		// Stop all containers labeled with the component name
		label := "component=" + cmpName
		dockerClient.StopContainers(label)

		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("When devfile exec command is executed", func() {

		It("should execute the given command successfully in the container", func() {
			utils.ExecCommand(context, cmpName)

			// Check to see if it's been updated
			containers := dockerClient.GetRunningContainersByCompAlias(cmpName, "runtime")
			Expect(len(containers)).To(Equal(1))

			listDir := dockerClient.ExecContainer(containers[0], "ls -la "+sourcePath)
			helper.MatchAllInOutput(listDir, []string{"blah.js"})
		})

		It("should error out when no command is given by the user", func() {
			utils.ExecWithoutCommand(context, cmpName)
		})

		It("should error out when a invalid command is given by the user", func() {
			utils.ExecWithInvalidCommand(context, cmpName, "docker")
		})
	})
})
