package devfile

import (
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile build-images command tests", func() {

	var commonVar helper.CommonVar

	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("using a devfile.yaml containing an Image component", func() {

		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-outerloop.yaml"), path.Join(commonVar.Context, "devfile.yaml"))
			helper.Cmd("odo", "create").ShouldPass()

		})
		It("should run odo build-images without push", func() {
			stdout := helper.Cmd("odo", "build-images").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			Expect(stdout).To(ContainSubstring("build -t quay.io/unknown-account/myimage -f " + commonVar.Context + "/Dockerfile " + commonVar.Context))
		})

		It("should run odo build-images --push", func() {
			stdout := helper.Cmd("odo", "build-images", "--push").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			Expect(stdout).To(ContainSubstring("build -t quay.io/unknown-account/myimage -f " + commonVar.Context + "/Dockerfile " + commonVar.Context))
			Expect(stdout).To(ContainSubstring("push quay.io/unknown-account/myimage"))
		})
	})

	When("using a devfile.yaml containing an Image component with Dockerfile args", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-outerloop-args.yaml"), path.Join(commonVar.Context, "devfile.yaml"))
			helper.Cmd("odo", "create").ShouldPass()
		})

		It("should use args to build image when running odo build-images", func() {
			stdout := helper.Cmd("odo", "build-images").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			Expect(stdout).To(ContainSubstring("--unknown-flag value"))
		})

	})
})
