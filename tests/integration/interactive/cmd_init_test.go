package interactive

import (
	"bytes"
	"fmt"

	"github.com/Netflix/go-expect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo login and logout command tests", func() {

	var commonVar helper.CommonVar

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

		Command := []string{"odo", "init"}
		output, err := helper.RunInteractive(commonVar, Command, func(c *expect.Console, output *bytes.Buffer) (bytes.Buffer, error) {
			//buf := new(bytes.Buffer)
			res, err := c.ExpectString("Select language")
			if err != nil {
				return *output, nil
			}
			fmt.Fprintln(output, res)
			c.SendLine("go")
			res, err = c.ExpectString("Select project type")
			if err != nil {
				return *output, nil
			}
			fmt.Fprintln(output, res)
			c.SendLine("\n")
			res, err = c.ExpectString("Which starter project do you want to use")
			if err != nil {
				return *output, nil
			}
			fmt.Fprintln(output, res)
			c.SendLine("\n")
			res, err = c.ExpectString("Enter component name")
			if err != nil {
				return *output, nil
			}
			fmt.Fprintln(output, res)
			c.SendLine("my-go-app")
			res, err = c.ExpectString("Your new component \"my-go-app\" is ready in the current directory.")
			if err != nil {
				return *output, err
			}
			fmt.Fprintln(output, res)
			return *output, nil
		})
		// output, err := helper.RunInteractive(commonVar)

		Expect(err).To(BeNil())
		Expect(output).To(ContainSubstring("Your new component \"my-go-app\" is ready in the current directory."))
		Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElements("devfile.yaml"))
	})
})
