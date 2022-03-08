//go:build linux || darwin || dragonfly || solaris || openbsd || netbsd || freebsd
// +build linux darwin dragonfly solaris openbsd netbsd freebsd

package interactive

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo dev interactive command tests", func() {

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

	When("directory is empty", func() {

		BeforeEach(func() {
			Expect(helper.ListFilesInDir(commonVar.Context)).To(HaveLen(0))
		})

		It("should error", func() {
			output := helper.Cmd("odo", "dev").ShouldFail().Err()
			Expect(output).To(ContainSubstring("this command cannot run in an empty directory"))

		})
	})

	When("directory is not empty", func() {

		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "python"), commonVar.Context)
			Expect(helper.ListFilesInDir(commonVar.Context)).To(
				SatisfyAll(
					HaveLen(2),
					ContainElements("requirements.txt", "wsgi.py")))
		})

		It("should run alizer to download devfile", func() {

			language := "python"
			_, _ = helper.RunInteractive([]string{"odo", "dev"},
				nil,
				func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Based on the files in the current directory odo detected")

					helper.ExpectString(ctx, fmt.Sprintf("Language: %s", language))

					helper.ExpectString(ctx, fmt.Sprintf("Project type: %s", language))

					helper.ExpectString(ctx,
						fmt.Sprintf("The devfile \"%s\" from the registry \"DefaultDevfileRegistry\" will be downloaded.", language))

					helper.ExpectString(ctx, "Is this correct")
					helper.SendLine(ctx, "\n")

					helper.ExpectString(ctx, "Select container for which you want to change configuration")
					helper.SendLine(ctx, "\n")

					helper.ExpectString(ctx, "Enter component name")
					helper.SendLine(ctx, "my-app")

					helper.ExpectString(ctx, "Press Ctrl+c to exit")
					ctx.StopCommand()
				})

			Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElements("devfile.yaml"))
		})
	})
})
