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
		It("should run odo build-images with push", func() {
			stdout := helper.Cmd("odo", "build-images").ShouldPass().Out()
			Expect(stdout).To(ContainSubstring("Successfully tagged quay.io/unknown-account/myimage:latest"))
		})

		It("should run odo build-images --push", func() {
			stderr := helper.Cmd("odo", "build-images", "--push").ShouldFail().Err()
			Expect(stderr).To(MatchRegexp("unauthorized"))
		})
	})

	When("using a devfile.yaml containing an Image component with Dockerfile args", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-outerloop-args.yaml"), path.Join(commonVar.Context, "devfile.yaml"))
			helper.Cmd("odo", "create").ShouldPass()
		})

		It("should use args to build image when running odo build-images", func() {
			stderr := helper.Cmd("odo", "build-images").ShouldFail().Err()
			Expect(stderr).To(ContainSubstring("unknown flag: --unknown-flag"))
		})

	})
})
