package docker

import (
	"path/filepath"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo docker devfile url pushtarget command tests", func() {
	var cmpName string
	dockerClient := helper.NewDockerRunner("docker")
	var globals helper.Globals

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		globals = helper.CommonBeforeEachDocker()
		cmpName = helper.RandString(6)
		helper.Chdir(globals.Context)
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

	// These tests require an active kube context *and* Docker daemon, so keeping them separate
	// from the other Docker URL tests which only require Docker.
	Context("Switching pushtarget", func() {
		It("switch from docker to kube, odo push should display warning", func() {
			var stdout string

			helper.CmdShouldPass("odo", "create", "nodejs", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), globals.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(globals.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create")

			helper.CmdShouldPass("odo", "preference", "set", "pushtarget", "kube", "-f")
			session := helper.CmdRunner("odo", "push", "--devfile", "devfile.yaml")
			stdout = string(session.Wait().Out.Contents())
			stderr := string(session.Wait().Err.Contents())
			Expect(stderr).To(ContainSubstring("Found a URL defined for Docker, but no valid URLs for Kubernetes."))
			Expect(stdout).To(ContainSubstring("Changes successfully pushed to component"))
		})

		It("switch from kube to docker, odo push should display warning", func() {
			var stdout string
			helper.CmdShouldPass("odo", "preference", "set", "pushtarget", "kube", "-f")
			helper.CmdShouldPass("odo", "create", "nodejs", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), globals.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(globals.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", "--host", "1.2.3.4.com", "--ingress")

			helper.CmdShouldPass("odo", "preference", "set", "pushtarget", "docker", "-f")
			session := helper.CmdRunner("odo", "push", "--devfile", "devfile.yaml")
			stdout = string(session.Wait().Out.Contents())
			stderr := string(session.Wait().Err.Contents())
			Expect(stderr).To(ContainSubstring("Found a URL defined for Kubernetes, but no valid URLs for Docker."))
			Expect(stdout).To(ContainSubstring("Changes successfully pushed to component"))
		})
	})

})
