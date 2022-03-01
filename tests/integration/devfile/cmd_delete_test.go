package devfile

import (
	. "github.com/onsi/ginkgo"
	"path"
	"path/filepath"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo devfile delete command tests", func() {
	var componentName string

	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		componentName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})
	When("a component is created", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-deploy-with-multiple-resources.yaml"), path.Join(commonVar.Context, "devfile.yaml"))
		})
		It("should not fail when odo delete component is run", func() {
			helper.Cmd("odo", "delete", "component", "-f").ShouldPass()
		})
		When("the component is pushed", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "push").ShouldPass()
			})
			It("should not fail when odo delete component is run", func() {
				helper.Cmd("odo", "delete", "component", "-f").ShouldPass()
			})
		})
		When("the component is deployed", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass()
			})
			It("should not fail when odo delete component is run", func() {
				helper.Cmd("odo", "delete", "component", "-f").ShouldPass()
			})
		})
	})
})
