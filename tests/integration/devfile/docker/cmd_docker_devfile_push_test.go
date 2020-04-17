package docker

import (
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo docker devfile push command tests", func() {
	var context string
	var currentWorkingDirectory string
	var cmpName string

	dockerClient := helper.NewDockerRunner("docker")

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
		cmpName = helper.RandString(6)
		helper.Chdir(context)
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
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

	Context("Verify devfile push works", func() {

		It("Check that odo push works with a devfile", func() {
			// Local devfile push requires experimental mode to be set and the pushtarget set to docker
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			helper.CmdShouldPass("odo", "preference", "set", "pushtarget", "docker")

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), context)
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, cmpName)

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml")
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// update devfile and push again
			helper.ReplaceString("devfile.yaml", "name: FOO", "name: BAR")
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml")
		})

		It("Check that odo push works with a devfile that has multiple containers", func() {
			// Local devfile push requires experimental mode to be set and the pushtarget set to docker
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			helper.CmdShouldPass("odo", "preference", "set", "pushtarget", "docker")

			// Springboot devfile references multiple containers
			helper.CmdShouldPass("odo", "create", "java-spring-boot", "--context", context, cmpName)

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml")
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// update devfile and push again
			helper.ReplaceString("devfile.yaml", "name: FOO", "name: BAR")
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml")
		})

		It("Check that odo push works with a devfile that has volumes defined", func() {
			// Local devfile push requires experimental mode to be set and the pushtarget set to docker
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			helper.CmdShouldPass("odo", "preference", "set", "pushtarget", "docker")

			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), context)
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
			// Local devfile push requires experimental mode to be set and the pushtarget set to docker
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			helper.CmdShouldPass("odo", "preference", "set", "pushtarget", "docker")

			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), context)
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

	})

})
