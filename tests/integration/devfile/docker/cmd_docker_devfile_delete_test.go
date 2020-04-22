package docker

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo docker devfile delete command tests", func() {
	var context string
	var currentWorkingDirectory string
	var cmpName string
	var projectDir = "/projectDir"
	var projectDirPath string

	dockerClient := helper.NewDockerRunner("docker")

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
		cmpName = helper.RandString(6)
		projectDirPath = context + projectDir
		helper.Chdir(context)
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		// Devfile commands require experimental mode to be set
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
		helper.CmdShouldPass("odo", "preference", "set", "pushtarget", "docker")
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {

		// Stop all containers labeled with the component name
		label := "component=" + cmpName
		dockerClient.StopContainers(label)

		dockerClient.RemoveVolumesByComponentAndType(cmpName, "projects")

		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")

	})

	Context("when docker devfile delete command is executed", func() {

		It("should delete the component created from the devfile and also the owned resources", func() {

			helper.CmdShouldPass("git", "clone", "https://github.com/che-samples/web-nodejs-sample.git", projectDirPath)
			helper.Chdir(projectDirPath)

			helper.CmdShouldPass("odo", "create", "nodejs", cmpName)

			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml")

			Expect(dockerClient.GetRunningContainersByLabel("component=" + cmpName)).To(HaveLen(1))

			Expect(dockerClient.ListVolumesOfComponentAndType(cmpName, "projects")).To(HaveLen(1))

			helper.CmdShouldPass("odo", "delete", "--devfile", "devfile.yaml", "-f")

			Expect(dockerClient.GetRunningContainersByLabel("component=" + cmpName)).To(HaveLen(0))

			Expect(dockerClient.ListVolumesOfComponentAndType(cmpName, "projects")).To(HaveLen(0))

		})
	})

	Context("when docker devfile delete command is executed with all flag", func() {

		It("should delete the component created from the devfile and also the env folder", func() {
			// helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), context)

			// helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, cmpName)

			helper.CmdShouldPass("git", "clone", "https://github.com/che-samples/web-nodejs-sample.git", projectDirPath)
			helper.Chdir(projectDirPath)

			helper.CmdShouldPass("odo", "create", "nodejs", cmpName)

			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml")

			Expect(dockerClient.GetRunningContainersByLabel("component=" + cmpName)).To(HaveLen(1))

			Expect(dockerClient.ListVolumesOfComponentAndType(cmpName, "projects")).To(HaveLen(1))

			helper.CmdShouldPass("odo", "delete", "--devfile", "devfile.yaml", "-f", "--all")

			Expect(dockerClient.GetRunningContainersByLabel("component=" + cmpName)).To(HaveLen(0))

			Expect(dockerClient.ListVolumesOfComponentAndType(cmpName, "projects")).To(HaveLen(0))

			files := helper.ListFilesInDir(projectDirPath)
			Expect(files).To(Not(ContainElement(".odo")))

		})
	})
})
