package devfile

import (
	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile exec command tests", func() {
	var cmpName string
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		cmpName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
		helper.SetDefaultDevfileRegistryAsStaging()
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("When devfile exec command is executed", func() {

		It("should execute the given command successfully in the container", func() {
			utils.ExecCommand(commonVar.Context, cmpName)
			podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
			listDir := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
			Expect(listDir).To(ContainSubstring("blah.js"))
		})

		It("should error out when no command is given by the user", func() {
			utils.ExecWithoutCommand(commonVar.Context, cmpName)
		})

		It("should error out when a invalid command is given by the user", func() {
			utils.ExecWithInvalidCommand(commonVar.Context, cmpName, "kube")
		})

		It("should error out when a component is not present or when a devfile flag is used", func() {
			utils.ExecCommandWithoutComponentAndDevfileFlag(commonVar.Context, cmpName)
		})
	})
})
