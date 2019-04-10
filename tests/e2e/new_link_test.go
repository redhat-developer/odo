package e2e

import (
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/e2e/helper"
)

// following command will tests in Describe section below in parallel (in 2 nodes)
// ginkgo -nodes=2 -focus="Example of a clean test" slowSpecThreshold=120 -randomizeAllSpecs  tests/e2e/
var _ = Describe("odoLinkE2e", func() {
	//new clean project and context for each test
	var project string
	var context string

	//  current directory and project (before eny test is run) so it can restored  after all testing is done
	var originalDir string

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is before every spec (It)
	var _ = BeforeEach(func() {
		project = helper.OcCreateRandProject()
		context = helper.CreateNewContext()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.OcDeleteProject(project)
		helper.DeleteDir(context)
	})

	var _ = Context("link", func() {

		// we will be testing components that are created from the current directory
		// switch to the clean context dir before each test
		var _ = JustBeforeEach(func() {
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})

		// go back to original directory after each test
		var _ = JustAfterEach(func() {
			helper.Chdir(originalDir)
		})

		var _ = Context("when --app flag is used", func() {
			It("create local openJDK component and push code and link to postgresql", func() {
				helper.CopyExample(filepath.Join("source", "openjdk-sb-postgresql"), context)
				importOpenJDKImage()

				helper.CmdShouldPass("odo create java sb-app --app jdklinktest")
				//TODO: verify that config was properly created
				helper.CmdShouldPass("odo push")

				helper.CmdShouldPass("odo service create dh-postgresql-apb --app jdklinktest --plan dev -p postgresql_user=luke -p postgresql_password=secret -p postgresql_database=my_data -p postgresql_version=9.6")
				waitForCmdOut("oc get serviceinstance -o name", 1, true, func(output string) bool {
					return strings.Contains(output, "dh-postgresql-apb")
				})
				outputList := helper.CmdShouldPass("odo service list")
				Expect(outputList).To(ContainSubstring("dh-postgresql-apb"))

				helper.CmdShouldPass("odo url create --port 8080")
				helper.CmdShouldPass("odo push")

				helper.CmdShouldPass("odo link dh-postgresql-apb -w --wait-for-target")
				waitForCmdOut("odo service list | sed 1d", 1, true, func(output string) bool {
					return strings.Contains(output, "dh-postgresql-apb") &&
						strings.Contains(output, "ProvisionedAndLinked")
				})

				routeURL := determineRouteURL()
				// Ping said URL
				responseStringMatchStatus := matchResponseSubString(routeURL, "Spring Boot", 30, 1)
				Expect(responseStringMatchStatus).Should(BeTrue())

				// Delete the component
				runCmdShouldPass("odo delete sb-app -f")
				// Delete the service
				runCmdShouldPass("odo service delete dh-postgresql-apb -f")

			})

		})

	})
})
