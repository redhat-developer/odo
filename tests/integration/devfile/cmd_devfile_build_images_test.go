package devfile

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
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
			helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-outerloop.yaml")).ShouldPass()
			helper.CreateLocalEnv(commonVar.Context, "aname", commonVar.Project)
		})
		It("should run odo build-images without push", func() {
			stdout := helper.Cmd("odo", "build-images").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			Expect(stdout).To(ContainSubstring("build -t quay.io/unknown-account/myimage -f " + filepath.Join(commonVar.Context, "Dockerfile ") + commonVar.Context))
		})

		It("should run odo build-images --push", func() {
			stdout := helper.Cmd("odo", "build-images", "--push").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			Expect(stdout).To(ContainSubstring("build -t quay.io/unknown-account/myimage -f " + filepath.Join(commonVar.Context, "Dockerfile ") + commonVar.Context))
			Expect(stdout).To(ContainSubstring("push quay.io/unknown-account/myimage"))
		})
	})

	When("using a devfile.yaml containing an Image component with Dockerfile args", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-outerloop-args.yaml")).ShouldPass()
			helper.CreateLocalEnv(commonVar.Context, "aname", commonVar.Project)
		})

		It("should use args to build image when running odo build-images", func() {
			stdout := helper.Cmd("odo", "build-images").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			Expect(stdout).To(ContainSubstring("--unknown-flag value"))
		})

	})
})
