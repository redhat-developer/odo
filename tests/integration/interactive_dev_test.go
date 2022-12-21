package integration

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
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

	Context("directory is not empty", func() {

		When("there is a match from Alizer", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "python"), commonVar.Context)
				Expect(helper.ListFilesInDir(commonVar.Context)).To(
					SatisfyAll(
						HaveLen(2),
						ContainElements("requirements.txt", "wsgi.py")))
			})
			It("should run alizer to download devfile successfully even with -v flag", func() {

				language := "Python"
				projectType := "Python"
				versionedDevfileName := "python:2.1.0"
				_, _ = helper.RunInteractive([]string{"odo", "dev", "--random-ports", "-v", "4"},
					nil,
					func(ctx helper.InteractiveContext) {
						helper.ExpectString(ctx, "Based on the files in the current directory odo detected")

						helper.ExpectString(ctx, fmt.Sprintf("Language: %s", language))

						helper.ExpectString(ctx, fmt.Sprintf("Project type: %s", projectType))

						helper.ExpectString(ctx,
							fmt.Sprintf("The devfile %q from the registry \"DefaultDevfileRegistry\" will be downloaded.", versionedDevfileName))

						helper.ExpectString(ctx, "Is this correct")
						helper.SendLine(ctx, "")

						helper.ExpectString(ctx, "Select container for which you want to change configuration")
						helper.SendLine(ctx, "")

						helper.ExpectString(ctx, "Enter component name")
						helper.SendLine(ctx, "my-app")

						helper.ExpectString(ctx, "[Ctrl+c] - Exit")
						ctx.StopCommand()
					})

				Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElements("devfile.yaml"))
			})

			It("should run alizer to download devfile", func() {

				language := "Python"
				projectType := "Python"
				versionedDevfileName := "python:2.1.0"
				_, _ = helper.RunInteractive([]string{"odo", "dev", "--random-ports"},
					nil,
					func(ctx helper.InteractiveContext) {
						helper.ExpectString(ctx, "Based on the files in the current directory odo detected")

						helper.ExpectString(ctx, fmt.Sprintf("Language: %s", language))

						helper.ExpectString(ctx, fmt.Sprintf("Project type: %s", projectType))

						helper.ExpectString(ctx,
							fmt.Sprintf("The devfile %q from the registry \"DefaultDevfileRegistry\" will be downloaded.", versionedDevfileName))

						helper.ExpectString(ctx, "Is this correct")
						helper.SendLine(ctx, "")

						helper.ExpectString(ctx, "Select container for which you want to change configuration")
						helper.SendLine(ctx, "")

						helper.ExpectString(ctx, "Enter component name")
						helper.SendLine(ctx, "my-app")

						helper.ExpectString(ctx, "[Ctrl+c] - Exit")
						ctx.StopCommand()
					})

				Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElements("devfile.yaml"))
			})

			It("should display welcoming messages first", func() {

				if os.Getenv("SKIP_WELCOMING_MESSAGES") == "true" {
					Skip("This is a Unix specific scenario, skipping")
				}

				language := "Python"
				projectType := "Python"
				versionedDevfileName := "python:2.1.0"
				output, _ := helper.RunInteractive([]string{"odo", "dev", "--random-ports"},
					// Setting verbosity level to 0, because we would be asserting the welcoming message is the first
					// message displayed to the end user. So we do not want any potential debug lines to be printed first.
					// Using envvars here (and not via the -v flag), because of https://github.com/redhat-developer/odo/issues/5513
					[]string{"ODO_LOG_LEVEL=0"},
					func(ctx helper.InteractiveContext) {
						helper.ExpectString(ctx, "Based on the files in the current directory odo detected")

						helper.ExpectString(ctx, fmt.Sprintf("Language: %s", language))

						helper.ExpectString(ctx, fmt.Sprintf("Project type: %s", projectType))

						helper.ExpectString(ctx,
							fmt.Sprintf("The devfile %q from the registry \"DefaultDevfileRegistry\" will be downloaded.", versionedDevfileName))

						helper.ExpectString(ctx, "Is this correct")
						helper.SendLine(ctx, "")

						helper.ExpectString(ctx, "Select container for which you want to change configuration")
						helper.SendLine(ctx, "")

						helper.ExpectString(ctx, "Enter component name")
						helper.SendLine(ctx, "my-app")

						helper.ExpectString(ctx, "[Ctrl+c] - Exit")
						ctx.StopCommand()
					})

				lines, err := helper.ExtractLines(output)
				Expect(err).To(BeNil())
				Expect(lines).To(Not(BeEmpty()))
				Expect(lines[1]).To(ContainSubstring("Dev mode ran, but no Devfile was found. Initializing a component in the current directory"))
			})
		})

		When("Alizer cannot determine a Devfile based on the current source code", func() {
			BeforeEach(func() {
				helper.CreateSimpleFile(commonVar.Context, "some-file-", ".ext")
			})

			It("should not fail but fallback to the interactive mode", func() {
				output, _ := helper.RunInteractive([]string{"odo", "dev", "--random-ports"}, nil, func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Could not determine a Devfile based on the files in the current directory")

					helper.ExpectString(ctx, "Select language")
					helper.SendLine(ctx, "Python")

					helper.ExpectString(ctx, "Select project type")
					helper.SendLine(ctx, "Python")

					helper.ExpectString(ctx, "Select version")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Select container for which you want to change configuration?")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Enter component name")
					helper.SendLine(ctx, "my-app")

					helper.ExpectString(ctx, "Building your application in container on cluster")
					ctx.StopCommand()
				})

				Expect(output).ShouldNot(ContainSubstring("Which starter project do you want to use"))
				Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElement("devfile.yaml"))
			})
		})
	})

	When("a component is bootstrapped", func() {

		var cmpName string

		BeforeEach(func() {
			cmpName = helper.RandString(6)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
		})

		It("should sync files when p is pressed", func() {
			_, _ = helper.RunInteractive([]string{"odo", "dev", "--random-ports", "--no-watch"},
				nil,
				func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "[p] - Manually apply")

					helper.PressKey(ctx, 'p')

					helper.ExpectString(ctx, "Pushing files")
					ctx.StopCommand()
				})
		})
	})
})
