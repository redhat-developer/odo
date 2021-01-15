package devfile

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile delete command tests", func() {
	const devfile = "devfile.yaml"
	var devfilePath string
	var componentName, invalidNamespace string

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

	Context("when devfile delete command is executed", func() {
		It("should not throw an error with an existing namespace when no component exists", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, componentName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			helper.CmdShouldPass("odo", "delete", "--project", commonVar.Project, "-f")
		})

		It("should delete the component created from the devfile and also the owned resources", func() {
			resourceTypes := []string{"deployments", "pods", "services", "ingress"}

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, componentName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io", "--port", "3000", "--ingress", "--context", commonVar.Context)

			if os.Getenv("KUBERNETES") != "true" {
				helper.CmdShouldPass("odo", "url", "create", "example-1", "--port", "3000", "--context", commonVar.Context)
				resourceTypes = append(resourceTypes, "routes")
			}

			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)

			helper.CmdShouldPass("odo", "delete", "--project", commonVar.Project, "-f")

			for _, resourceType := range resourceTypes {
				commonVar.CliRunner.WaitAndCheckForExistence(resourceType, commonVar.Project, 1)
			}
		})

		It("should delete the component created from the devfile and also the env and odo folders and the odo-index-file.json file with all flag", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, componentName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)

			helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io", "--port", "3000", "--ingress", "--context", commonVar.Context)

			if os.Getenv("KUBERNETES") != "true" {
				helper.CmdShouldPass("odo", "url", "create", "example-1", "--port", "3000")
			}

			helper.CmdShouldPass("odo", "delete", "--project", commonVar.Project, "-f", "--all")

			commonVar.CliRunner.WaitAndCheckForExistence("deployments", commonVar.Project, 1)

			files := helper.ListFilesInDir(commonVar.Context)
			Expect(files).To(Not(ContainElement(".odo")))
			Expect(files).To(Not(ContainElement("devfile.yaml")))
		})

		It("should execute preStop events if present", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, componentName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-valid-events.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)

			output := helper.CmdShouldPass("odo", "delete", "--project", commonVar.Project, "-f")
			helper.MatchAllInOutput(output, []string{
				fmt.Sprintf("Executing preStop event commands for component %s", componentName),
				"Executing myprestop command",
				"Executing secondprestop command",
				"Executing thirdprestop command",
			})

		})

		It("should error out on devfile flag", func() {
			helper.CmdShouldFail("odo", "delete", "--devfile", "invalid.yaml")
		})
	})

	Context("when the project doesn't exist", func() {
		JustBeforeEach(func() {
			invalidNamespace = "garbage"
		})

		It("should let the user delete the local config files with -a flag", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", invalidNamespace, componentName)
			utils.DeleteLocalConfig("delete")
		})

		It("should let the user delete the local config files with -a and -project flags", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", invalidNamespace, componentName)
			utils.DeleteLocalConfig("delete", "--project", invalidNamespace)
		})
	})

	Context("When devfile exists not in user's working directory and user specify the devfile path via --devfile", func() {
		JustBeforeEach(func() {
			newContext := path.Join(commonVar.Context, "newContext")
			devfilePath = filepath.Join(newContext, devfile)
			helper.MakeDir(newContext)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), devfilePath)
		})

		It("should successfully delete the devfile as its not present in root on delete", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--devfile", devfilePath)
			// devfile was copied to top level
			Expect(helper.VerifyFileExists(path.Join(commonVar.Context, devfile))).To(BeTrue())
			helper.CmdShouldPass("odo", "delete", "--all", "-f")
			Expect(helper.VerifyFileExists(path.Join(commonVar.Context, devfile))).To(BeFalse())
		})

		It("should not delete the devfile if its already present", func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), path.Join(commonVar.Context, devfile))
			helper.CmdShouldPass("odo", "create", "nodejs")
			// devfile was copied to top level
			Expect(helper.VerifyFileExists(path.Join(commonVar.Context, devfile))).To(BeTrue())
			helper.CmdShouldPass("odo", "delete", "--all", "-f")
			Expect(helper.VerifyFileExists(path.Join(commonVar.Context, devfile))).To(BeTrue())
		})

	})

})
