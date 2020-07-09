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

	var fakeVolumes []string

	dockerClient := helper.NewDockerRunner("docker")

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
		cmpName = helper.RandString(6)
		helper.Chdir(context)
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		// Devfile commands require experimental mode to be set
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
		helper.CmdShouldPass("odo", "preference", "set", "pushtarget", "docker")

		// With our docker delete code, we want to avoid deleting volumes that we didn't create. So
		// next we create a set of fake volumes, none of which should be deleted (eg they are out of scope) by any of the tests.

		// 1) Create volume with fake component but valid type
		volname1 := cmpName + "-fakevol1"
		fakeVolumes = append(fakeVolumes, volname1)
		dockerClient.CreateVolume(volname1, []string{"component=fake", "type=projects"})

		// 2) Create volume with fake component but valid storage
		volname2 := cmpName + "-fakevol2"
		fakeVolumes = append(fakeVolumes, volname2)
		dockerClient.CreateVolume(volname2, []string{"component=fake", "storage-name=" + volname2})

		// 3) Create volume with real component but neither valid source ("type") nor valid storage
		volname3 := cmpName + "-fakevol3"
		fakeVolumes = append(fakeVolumes, volname3)
		dockerClient.CreateVolume(volname3, []string{"component=" + cmpName, "type=not-projects", "storage-name-fake=" + volname3})

	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {

		// Ensure that our fake volumes all still exist, then clean them up.
		for _, fakeVolume := range fakeVolumes {
			Expect(dockerClient.VolumeExists(fakeVolume)).To(Equal(true))
			dockerClient.RemoveVolumeByName(fakeVolume)
		}
		fakeVolumes = []string{}

		// Stop all containers labeled with the component name
		label := "component=" + cmpName
		dockerClient.StopContainers(label)

		dockerClient.RemoveVolumesByComponent(cmpName)

		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")

	})

	Context("when docker devfile delete command is executed", func() {

		It("should delete the component created from the devfile and also the owned resources", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push")

			Expect(dockerClient.GetRunningContainersByLabel("component=" + cmpName)).To(HaveLen(1))

			Expect(dockerClient.GetSourceAndStorageVolumesByComponent(cmpName)).To(HaveLen(1))

			helper.CmdShouldPass("odo", "delete", "-f")

			Expect(dockerClient.GetRunningContainersByLabel("component=" + cmpName)).To(HaveLen(0))

			Expect(dockerClient.GetSourceAndStorageVolumesByComponent(cmpName)).To(HaveLen(0))

		})

		It("should delete all the mounted volume types listed in the devfile", func() {

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

			Expect(dockerClient.GetSourceAndStorageVolumesByComponent(cmpName)).To(HaveLen(3))

			helper.CmdShouldPass("odo", "delete", "-f")

			Expect(dockerClient.GetRunningContainersByLabel("component=" + cmpName)).To(HaveLen(0))

			Expect(dockerClient.GetSourceAndStorageVolumesByComponent(cmpName)).To(HaveLen(0))

		})

	})

	Context("when no component exists", func() {

		It("should not throw an error", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", cmpName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "delete", "-f")
		})
	})

	Context("when docker devfile delete command is executed with all flag", func() {

		It("should delete the component created from the devfile and also the env folder", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push")

			Expect(dockerClient.GetRunningContainersByLabel("component=" + cmpName)).To(HaveLen(1))

			Expect(dockerClient.GetSourceAndStorageVolumesByComponent(cmpName)).To(HaveLen(1))

			helper.CmdShouldPass("odo", "delete", "-f", "--all")

			Expect(dockerClient.GetRunningContainersByLabel("component=" + cmpName)).To(HaveLen(0))

			Expect(dockerClient.GetSourceAndStorageVolumesByComponent(cmpName)).To(HaveLen(0))

			files := helper.ListFilesInDir(context)
			Expect(files).To(Not(ContainElement(".odo")))
			Expect(files).To(Not(ContainElement("devfile.yaml")))
		})
	})
})
