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

var _ = Describe("odo docker devfile watch command tests", func() {
	var context, currentWorkingDirectory, cmpName, projectDirPath string
	var projectDir = "/projectDir"

	dockerClient := helper.NewDockerRunner("docker")

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
		projectDirPath = context + projectDir
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

		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("when executing odo watch after odo push", func() {
		It("should listen for file changes", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/che-samples/web-nodejs-sample.git", projectDirPath)
			helper.Chdir(projectDirPath)

			helper.CmdShouldPass("odo", "create", "nodejs", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), projectDirPath)

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml")
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// odo watch and validate
			utils.OdoWatch(cmpName, "", projectDirPath, []string{"Executing devbuild command", "Executing devrun command"}, dockerClient, "docker")
		})
	})

	Context("when executing odo watch after odo push with custom commands", func() {
		It("should listen for file changes", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/che-samples/web-nodejs-sample.git", projectDirPath)
			helper.Chdir(projectDirPath)

			helper.CmdShouldPass("odo", "create", "nodejs", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), projectDirPath)

			output := helper.CmdShouldPass("odo", "push", "--build-command", "build", "--run-command", "run", "--devfile", "devfile.yaml")
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// odo watch and validate
			utils.OdoWatch(cmpName, "", projectDirPath, []string{"Executing build command", "Executing run command"}, dockerClient, "docker")
		})
	})

})
