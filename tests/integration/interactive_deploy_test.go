package integration

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo deploy interactive command tests", func() {

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

	When("directory is not empty", func() {

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
			devfileName := "python"
			output, err := helper.RunInteractive([]string{"odo", "deploy", "-v", "4"},
				nil,
				func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Based on the files in the current directory odo detected")

					helper.ExpectString(ctx, fmt.Sprintf("Language: %s", language))

					helper.ExpectString(ctx, fmt.Sprintf("Project type: %s", projectType))

					helper.ExpectString(ctx,
						fmt.Sprintf("The devfile \"%s\" from the registry \"DefaultDevfileRegistry\" will be downloaded.", devfileName))

					helper.ExpectString(ctx, "Is this correct")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Select container for which you want to change configuration")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Enter component name")
					helper.SendLine(ctx, "my-app")

					helper.ExpectString(ctx, "no default deploy command found in devfile")
				})

			Expect(err).To(Not(BeNil()))
			Expect(output).To(ContainSubstring("no default deploy command found in devfile"))
			Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElements("devfile.yaml"))
		})

		It("should run alizer to download devfile", func() {

			language := "Python"
			projectType := "Python"
			devfileName := "python"
			output, err := helper.RunInteractive([]string{"odo", "deploy"},
				nil,
				func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Based on the files in the current directory odo detected")

					helper.ExpectString(ctx, fmt.Sprintf("Language: %s", language))

					helper.ExpectString(ctx, fmt.Sprintf("Project type: %s", projectType))

					helper.ExpectString(ctx,
						fmt.Sprintf("The devfile \"%s\" from the registry \"DefaultDevfileRegistry\" will be downloaded.", devfileName))

					helper.ExpectString(ctx, "Is this correct")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Select container for which you want to change configuration")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Enter component name")
					helper.SendLine(ctx, "my-app")

					helper.ExpectString(ctx, "no default deploy command found in devfile")
				})

			Expect(err).To(Not(BeNil()))
			Expect(output).To(ContainSubstring("no default deploy command found in devfile"))
			Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElements("devfile.yaml"))
		})

		It("should display welcoming messages first", func() {

			if os.Getenv("SKIP_WELCOMING_MESSAGES") == "true" {
				Skip("This is a Unix specific scenario, skipping")
			}

			language := "Python"
			projectType := "Python"
			devfileName := "python"
			output, err := helper.RunInteractive([]string{"odo", "deploy"},
				// Setting verbosity level to 0, because we would be asserting the welcoming message is the first
				// message displayed to the end user. So we do not want any potential debug lines to be printed first.
				// Using envvars here (and not via the -v flag), because of https://github.com/redhat-developer/odo/issues/5513
				[]string{"ODO_LOG_LEVEL=0"},
				func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Based on the files in the current directory odo detected")

					helper.ExpectString(ctx, fmt.Sprintf("Language: %s", language))

					helper.ExpectString(ctx, fmt.Sprintf("Project type: %s", projectType))

					helper.ExpectString(ctx,
						fmt.Sprintf("The devfile \"%s\" from the registry \"DefaultDevfileRegistry\" will be downloaded.", devfileName))

					helper.ExpectString(ctx, "Is this correct")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Select container for which you want to change configuration")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Enter component name")
					helper.SendLine(ctx, "my-app")

					helper.ExpectString(ctx, "no default deploy command found in devfile")
				})

			Expect(err).To(Not(BeNil()))
			// Make sure it also displays welcoming messages first
			lines, err := helper.ExtractLines(output)
			Expect(err).To(BeNil())
			Expect(lines).To(Not(BeEmpty()))
			Expect(lines[1]).To(ContainSubstring("Deploy mode ran, but no Devfile was found. Initializing a component in the current directory"))
		})
	})
})
