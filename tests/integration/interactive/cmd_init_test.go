//go:build linux || darwin || dragonfly || solaris || openbsd || netbsd || freebsd
// +build linux darwin dragonfly solaris openbsd netbsd freebsd

package interactive

import (
	"bytes"
	"fmt"

	"github.com/Netflix/go-expect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo init interactive command tests", func() {

	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	It("should download correct devfile", func() {

		command := []string{"odo", "init"}
		output, err := helper.RunInteractive(command, func(c *expect.Console, output *bytes.Buffer) {

			res := helper.ExpectString(c, "Select language")
			fmt.Fprintln(output, res)
			helper.SendLine(c, "go")

			res = helper.ExpectString(c, "Select project type")
			fmt.Fprintln(output, res)
			helper.SendLine(c, "\n")

			res = helper.ExpectString(c, "Which starter project do you want to use")
			fmt.Fprintln(output, res)
			helper.SendLine(c, "\n")

			res = helper.ExpectString(c, "Enter component name")
			fmt.Fprintln(output, res)
			helper.SendLine(c, "my-go-app")

			res = helper.ExpectString(c, "Your new component \"my-go-app\" is ready in the current directory.")
			fmt.Fprintln(output, res)

		})

		Expect(err).To(BeNil())
		Expect(output).To(ContainSubstring("Your new component \"my-go-app\" is ready in the current directory."))
		Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElements("devfile.yaml"))
	})
})
