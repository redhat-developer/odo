package integration

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/redhat-developer/odo/pkg/odo/cli/messages"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	odolog "github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/version"
	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo init interactive command tests", func() {

	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach(helper.SetupClusterFalse)
		helper.Chdir(commonVar.Context)

		// We make EXPLICITLY sure that we are outputting with NO COLOR
		// this is because in some cases we are comparing the output with a colorized one
		os.Setenv("NO_COLOR", "true")
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	It("should not fail when using -v flag", func() {
		command := []string{"odo", "init", "-v", "4"}
		output, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {

			By("showing the interactive mode notice message", func() {
				helper.ExpectString(ctx, messages.InteractiveModeEnabled)
			})

			helper.ExpectString(ctx, "Select language")
			helper.SendLine(ctx, "Go")

			helper.ExpectString(ctx, "Select project type")
			helper.SendLine(ctx, "")

			helper.ExpectString(ctx, "Which starter project do you want to use")
			helper.SendLine(ctx, "")

			helper.ExpectString(ctx, "Enter component name")
			helper.SendLine(ctx, "my-go-app")

			helper.ExpectString(ctx, "Your new component 'my-go-app' is ready in the current directory")

		})
		Expect(err).To(BeNil())
		Expect(output).To(ContainSubstring("Your new component 'my-go-app' is ready in the current directory"))
		Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElements("devfile.yaml"))
	})

	It("should ask to re-enter the component name when an invalid value is passed", func() {
		command := []string{"odo", "init"}
		_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {

			helper.ExpectString(ctx, "Select language")
			helper.SendLine(ctx, "Go")

			helper.ExpectString(ctx, "Select project type")
			helper.SendLine(ctx, "")

			helper.ExpectString(ctx, "Which starter project do you want to use")
			helper.SendLine(ctx, "")

			helper.ExpectString(ctx, "Enter component name")
			helper.SendLine(ctx, "myapp-<script>alert('Injected!');</script>")

			helper.ExpectString(ctx, "name \"myapp-<script>alert('Injected!');</script>\" is not valid, name should conform the following requirements")

			helper.ExpectString(ctx, "Enter component name")
			helper.SendLine(ctx, "my-go-app")

			helper.ExpectString(ctx, "Your new component 'my-go-app' is ready in the current directory")
		})
		Expect(err).To(BeNil())
		Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElements("devfile.yaml"))
	})

	It("should download correct devfile", func() {

		command := []string{"odo", "init"}
		output, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {

			By("showing the interactive mode notice message", func() {
				helper.ExpectString(ctx, messages.InteractiveModeEnabled)
			})

			helper.ExpectString(ctx, "Select language")
			helper.SendLine(ctx, "Go")

			helper.ExpectString(ctx, "Select project type")
			helper.SendLine(ctx, "")

			helper.ExpectString(ctx, "Which starter project do you want to use")
			helper.SendLine(ctx, "")

			helper.ExpectString(ctx, "Enter component name")
			helper.SendLine(ctx, "my-go-app")

			helper.ExpectString(ctx, "Your new component 'my-go-app' is ready in the current directory")

		})

		Expect(err).To(BeNil())
		Expect(output).To(ContainSubstring("Your new component 'my-go-app' is ready in the current directory"))
		Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElements("devfile.yaml"))
	})

	Describe("displaying welcoming messages", func() {

		// testFunc is a function that returns a `Tester` function (intended to be used via `helper.RunInteractive`),
		// which first expects all messages in `welcomingMsgs` to be read from the console,
		// then runs an `additionalTester` and finally expects the asking of a component name
		// (based on the `language` specified)
		testFunc := func(language string, welcomingMsgs []string, additionalTester helper.Tester) helper.Tester {
			return func(ctx helper.InteractiveContext) {
				for _, msg := range welcomingMsgs {
					helper.ExpectString(ctx, msg)
				}

				if additionalTester != nil {
					additionalTester(ctx)
				}

				helper.ExpectString(ctx, "Enter component name")
				helper.SendLine(ctx, fmt.Sprintf("my-%s-app", language))

				helper.ExpectString(ctx,
					fmt.Sprintf("Your new component 'my-%s-app' is ready in the current directory", language))
			}
		}

		assertBehavior := func(language string, output string, err error, msgs []string, additionalAsserter func()) {
			Expect(err).To(BeNil())

			lines, err := helper.ExtractLines(output)
			if err != nil {
				log.Fatal(err)
			}
			Expect(len(lines)).To(BeNumerically(">", len(msgs)))
			Expect(lines[0:len(msgs)]).To(Equal(msgs))
			Expect(lines).To(
				ContainElement(fmt.Sprintf("Your new component 'my-%s-app' is ready in the current directory", strings.ToLower(language))))

			Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElements("devfile.yaml"))

			if additionalAsserter != nil {
				additionalAsserter()
			}
		}

		testRunner := func(language string, welcomingMsgs []string, tester helper.Tester) (string, error) {
			command := []string{"odo", "init"}
			return helper.RunInteractive(command,
				// Setting verbosity level to 0, because we would be asserting the welcoming message is the first
				// message displayed to the end user. So we do not want any potential debug lines to be printed first.
				// Using envvars here (and not via the -v flag), because of https://github.com/redhat-developer/odo/issues/5513
				[]string{"ODO_LOG_LEVEL=0"},
				testFunc(strings.ToLower(language), welcomingMsgs, tester))
		}

		When("directory is empty", func() {

			BeforeEach(func() {
				Expect(helper.ListFilesInDir(commonVar.Context)).To(HaveLen(0))
			})

			It("should display appropriate welcoming messages", func() {

				if os.Getenv("SKIP_WELCOMING_MESSAGES") == "true" {
					Skip("This is a Unix specific scenario, skipping")
				}

				language := "java"

				// The first output is welcoming message / paragraph / banner output
				welcomingMsgs := strings.Split(odolog.Stitle(messages.InitializingNewComponent, messages.NoSourceCodeDetected, "odo version: "+version.VERSION), "\n")

				output, err := testRunner(language, welcomingMsgs, func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Select language")
					helper.SendLine(ctx, language)

					helper.ExpectString(ctx, "Select project type")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Which starter project do you want to use")
					helper.SendLine(ctx, "")
				})

				assertBehavior(strings.ToLower(language), output, err, welcomingMsgs, nil)
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

			It("should display appropriate welcoming messages", func() {

				if os.Getenv("SKIP_WELCOMING_MESSAGES") == "true" {
					Skip("This is a Unix specific scenario, skipping")
				}

				language := "Python"
				projectType := "Flask"
				devfileName := "python"
				welcomingMsgs := strings.Split(odolog.Stitle(messages.InitializingNewComponent, messages.SourceCodeDetected, "odo version: "+version.VERSION), "\n")

				output, err := testRunner(language, welcomingMsgs, func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Based on the files in the current directory odo detected")

					helper.ExpectString(ctx, fmt.Sprintf("Language: %s", language))

					helper.ExpectString(ctx, fmt.Sprintf("Project type: %s", projectType))

					helper.ExpectString(ctx,
						fmt.Sprintf("The devfile \"%s\" from the registry \"DefaultDevfileRegistry\" will be downloaded.", devfileName))

					helper.ExpectString(ctx, "Is this correct")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Select container for which you want to change configuration")
					helper.SendLine(ctx, "")
				})

				assertBehavior(strings.ToLower(language), output, err, welcomingMsgs, func() {
					// Make sure the original source code files are still present
					Expect(helper.ListFilesInDir(commonVar.Context)).To(
						SatisfyAll(
							HaveLen(3),
							ContainElements("devfile.yaml", "requirements.txt", "wsgi.py")))
				})
			})
		})

		When("alizer detection of javascript name", func() {

			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				Expect(helper.ListFilesInDir(commonVar.Context)).To(
					SatisfyAll(
						HaveLen(3),
						ContainElements("Dockerfile", "package.json", "server.js")))
			})

			It("should display node-echo name", func() {
				language := "javascript"
				projectType := "nodejs"
				projectName := "node-echo"

				output, err := helper.RunInteractive([]string{"odo", "init"}, nil, func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Based on the files in the current directory odo detected")

					helper.ExpectString(ctx, fmt.Sprintf("Language: %s", language))

					helper.ExpectString(ctx, fmt.Sprintf("Project type: %s", projectType))

					helper.ExpectString(ctx,
						fmt.Sprintf("The devfile \"%s\" from the registry \"DefaultDevfileRegistry\" will be downloaded.", projectType))

					helper.ExpectString(ctx, "Is this correct")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Select container for which you want to change configuration")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Enter component name")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, fmt.Sprintf("Your new component '%s' is ready in the current directory", projectName))
				})
				Expect(err).To(BeNil())

				lines, err := helper.ExtractLines(output)
				Expect(err).To(BeNil())
				Expect(len(lines)).To(BeNumerically(">", 2))
				Expect(lines[len(lines)-1]).To(Equal(fmt.Sprintf("Your new component '%s' is ready in the current directory", projectName)))

			})
			It("should ask to re-enter the component name if invalid value is passed by the user", func() {
				language := "javascript"
				projectType := "nodejs"
				projectName := "node-echo"

				_, err := helper.RunInteractive([]string{"odo", "init"}, nil, func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Based on the files in the current directory odo detected")

					helper.ExpectString(ctx, fmt.Sprintf("Language: %s", language))

					helper.ExpectString(ctx, fmt.Sprintf("Project type: %s", projectType))

					helper.ExpectString(ctx,
						fmt.Sprintf("The devfile \"%s\" from the registry \"DefaultDevfileRegistry\" will be downloaded.", projectType))

					helper.ExpectString(ctx, "Is this correct")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Select container for which you want to change configuration")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Enter component name")
					helper.SendLine(ctx, "myapp-<script>alert('Injected!');</script>")

					helper.ExpectString(ctx, "name \"myapp-<script>alert('Injected!');</script>\" is not valid, name should conform the following requirements")

					helper.ExpectString(ctx, "Enter component name")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, fmt.Sprintf("Your new component '%s' is ready in the current directory", projectName))
				})
				Expect(err).To(BeNil())
				Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElements("devfile.yaml"))
			})
		})
	})

	It("should start downloading starter project only after all interactive questions have been asked", func() {

		output, err := helper.RunInteractive([]string{"odo", "init"}, nil, func(ctx helper.InteractiveContext) {

			helper.ExpectString(ctx, "Select language")
			helper.SendLine(ctx, "dotnet")

			helper.ExpectString(ctx, "Select project type")
			helper.SendLine(ctx, "")

			helper.ExpectString(ctx, "Which starter project do you want to use")
			helper.SendLine(ctx, "")

			helper.ExpectString(ctx, "Enter component name")
			helper.SendLine(ctx, "my-dotnet-app")

			helper.ExpectString(ctx, "Your new component 'my-dotnet-app' is ready in the current directory")
		})

		Expect(err).To(BeNil())

		lines, err := helper.ExtractLines(output)
		Expect(err).To(BeNil())
		Expect(len(lines)).To(BeNumerically(">", 2))
		Expect(lines[len(lines)-1]).To(Equal("Your new component 'my-dotnet-app' is ready in the current directory"))

		componentNameQuestionIdx, ok := helper.FindFirstElementIndexMatchingRegExp(lines, ".*Enter component name:.*")
		Expect(ok).To(BeTrue())
		starterProjectDownloadActionIdx, found := helper.FindFirstElementIndexMatchingRegExp(lines,
			".*Downloading starter project \"([^\\s]+)\" \\[.*")
		Expect(found).To(BeTrue())
		Expect(starterProjectDownloadActionIdx).To(SatisfyAll(
			Not(BeZero()),
			// #5495: component name question should be displayed before starter project is actually downloaded
			BeNumerically(">", componentNameQuestionIdx),
		), "Action 'Downloading starter project' should have been displayed after the last interactive question ('Enter component name')")

		Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElements("devfile.yaml"))
	})
})
