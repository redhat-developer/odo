package interactive

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo login and logout command tests", func() {

	var commonVar helper.CommonVar
	var interVar helper.Interactive

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
	})

	// // Clean up after the test
	// // This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})
	It("should download correct devfile", func() {
		interVar.Command = []string{"odo", "init"}
		interVar.ExpectFromtty = append(interVar.ExpectFromtty, "Select language")
		interVar.SendOntty = append(interVar.SendOntty, "go")
		interVar.ExpectFromtty = append(interVar.ExpectFromtty, "Select project type")
		interVar.SendOntty = append(interVar.SendOntty, "\n")
		interVar.ExpectFromtty = append(interVar.ExpectFromtty, "Which starter project do you want to use")
		interVar.SendOntty = append(interVar.SendOntty, "\n")
		interVar.ExpectFromtty = append(interVar.ExpectFromtty, "Enter component name")
		interVar.SendOntty = append(interVar.SendOntty, "my-go-app")
		interVar.ExpectFromtty = append(interVar.ExpectFromtty, "Your new component \"my-go-app\" is ready in the current directory.")
		output, err := helper.RunInteractive(commonVar, interVar)
		// output, err := helper.RunInteractive(commonVar)

		Expect(err).To(BeNil())
		Expect(output).To(ContainSubstring("Your new component \"my-go-app\" is ready in the current directory."))
		Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElements("devfile.yaml"))
	})
})
