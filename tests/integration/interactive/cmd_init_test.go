//go:build linux || darwin || dragonfly || solaris || openbsd || netbsd || freebsd
// +build linux darwin dragonfly solaris openbsd netbsd freebsd

package interactive

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

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

	Describe("displaying welcoming messages", func() {

		// testFunc is a function that returns a `Tester` function (intended to be used via `helper.RunInteractive`),
		// which first expects all messages in `welcomingMsgs` to be read from the console,
		// then runs an `additionalTester` and finally expects the asking of a component name
		// (based on the `language` specified)
		testFunc := func(language string, welcomingMsgs []string, additionalTester helper.Tester) helper.Tester {
			return func(c *expect.Console, output *bytes.Buffer) {
				var res string
				for _, msg := range welcomingMsgs {
					res = helper.ExpectString(c, msg)
					fmt.Fprint(output, res)
				}

				if additionalTester != nil {
					additionalTester(c, output)
				}

				res = helper.ExpectString(c, "Enter component name")
				fmt.Fprintln(output, res)
				helper.SendLine(c, fmt.Sprintf("my-%s-app", language))

				res = helper.ExpectString(c,
					fmt.Sprintf("Your new component \"my-%s-app\" is ready in the current directory.", language))
				fmt.Fprintln(output, res)
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
				ContainElement(fmt.Sprintf("Your new component \"my-%s-app\" is ready in the current directory.", language)))

			Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElements("devfile.yaml"))

			if additionalAsserter != nil {
				additionalAsserter()
			}
		}

		testRunner := func(language string, welcomingMsgs []string, tester helper.Tester) (string, error) {
			command := []string{"odo", "init"}
			return helper.RunInteractive(command, testFunc(language, welcomingMsgs, tester))
		}

		When("directory is empty", func() {

			BeforeEach(func() {
				Expect(helper.ListFilesInDir(commonVar.Context)).To(HaveLen(0))
			})

			It("should display appropriate welcoming messages", func() {
				language := "java"
				welcomingMsgs := []string{
					"The current directory is empty. odo will help you start a new project.",
				}

				output, err := testRunner(language, welcomingMsgs, func(c *expect.Console, output *bytes.Buffer) {
					res := helper.ExpectString(c, "Select language")
					fmt.Fprintln(output, res)
					helper.SendLine(c, language)

					res = helper.ExpectString(c, "Select project type")
					fmt.Fprintln(output, res)
					helper.SendLine(c, "\n")

					res = helper.ExpectString(c, "Which starter project do you want to use")
					fmt.Fprintln(output, res)
					helper.SendLine(c, "\n")
				})

				assertBehavior(language, output, err, welcomingMsgs, nil)
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
				language := "python"
				welcomingMsgs := []string{
					"The current directory already contains source code. " +
						"odo will try to autodetect the language and project type in order to select the best suited Devfile for your project.",
				}

				output, err := testRunner(language, welcomingMsgs, func(c *expect.Console, output *bytes.Buffer) {
					res := helper.ExpectString(c, "Based on the files in the current directory odo detected")
					fmt.Fprintln(output, res)

					res = helper.ExpectString(c, fmt.Sprintf("Language: %s", language))
					fmt.Fprintln(output, res)

					res = helper.ExpectString(c, fmt.Sprintf("Project type: %s", language))
					fmt.Fprintln(output, res)

					res = helper.ExpectString(c,
						fmt.Sprintf("The devfile \"%s\" from the registry \"DefaultDevfileRegistry\" will be downloaded.", language))
					fmt.Fprintln(output, res)

					res = helper.ExpectString(c, "Is this correct")
					fmt.Fprintln(output, res)
					helper.SendLine(c, "\n")

					res = helper.ExpectString(c, "Select container for which you want to change configuration")
					fmt.Fprintln(output, res)
					helper.SendLine(c, "\n")
				})

				assertBehavior(language, output, err, welcomingMsgs, func() {
					// Make sure the original source code files are still present
					Expect(helper.ListFilesInDir(commonVar.Context)).To(
						SatisfyAll(
							HaveLen(3),
							ContainElements("devfile.yaml", "requirements.txt", "wsgi.py")))
				})
			})
		})
	})
})
