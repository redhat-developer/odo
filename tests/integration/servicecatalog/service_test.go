package integration

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odoServiceE2e", func() {
	//new clean project and context for each test
	var project string
	var context string
	var app string

	//  current directory and project (before eny test is run) so it can restored  after all testing is done
	var originalDir string
	var oc helper.OcRunner
	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is before every spec (It)
	BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		oc = helper.NewOcRunner("oc")
	})

	Context("when running help for service command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "service", "-h")
			Expect(appHelp).To(ContainSubstring("Perform service catalog operations"))
		})
	})

	Context("When creating with a spring boot application", func() {
		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
			project = helper.CreateRandProject()
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})
		JustAfterEach(func() {
			helper.DeleteProject(project)
			helper.DeleteDir(context)
			helper.Chdir(originalDir)
			os.Unsetenv("GLOBALODOCONFIG")
		})
		It("should be able to create postgresql and link it with springboot", func() {
			oc.ImportJavaIsToNspace(project)
			helper.CopyExample(filepath.Join("source", "openjdk-sb-postgresql"), context)

			// Local config needs to be present in order to create service https://github.com/openshift/odo/issues/1602
			helper.CmdShouldPass("odo", "create", "java", "sb-app", "--project", project)

			helper.CmdShouldPass("odo", "service", "create", "dh-postgresql-apb", "--project", project, "--plan", "dev",
				"-p", "postgresql_user=luke", "-p", "postgresql_password=secret",
				"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6")
			ocArgs := []string{"get", "serviceinstance", "-o", "name", "-n", project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "dh-postgresql-apb")
			})

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080")

			// push removes link, this is why link needs to be run alaways after the push https://github.com/openshift/odo/issues/1596
			helper.CmdShouldPass("odo", "push", "-v", "4")

			helper.CmdShouldPass("odo", "link", "dh-postgresql-apb", "-w", "--wait-for-target")

			odoArgs := []string{"service", "list"}
			helper.WaitForCmdOut("odo", odoArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "dh-postgresql-apb") &&
					strings.Contains(output, "ProvisionedAndLinked")
			})

			routeURL := helper.DetermineRouteURL("")

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Spring Boot", 90, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "delete", "sb-app", "-f")

			// Delete the service
			helper.CmdShouldPass("odo", "service", "delete", "dh-postgresql-apb", "-f")
		})
	})

	// TODO: auth issue, we need to find a proper way how to test it without requiring cluster admin privileges

	// Context("odo hides a hidden service in service catalog", func() {
	// 	It("not show a hidden service in the catalog", func() {
	// 		runCmdShouldPass("oc apply -f https://github.com/openshift/library/raw/master/official/sso/templates/sso72-https.json -n openshift")
	// 		outputErr := runCmdShouldFail("odo catalog search service sso72-https")
	// 		Expect(outputErr).To(ContainSubstring("No service matched the query: sso72-https"))
	// 	})
	// })

	Context("When working from outside a component dir", func() {
		JustBeforeEach(func() {
			project = helper.CreateRandProject()
			context = helper.CreateNewContext()
			app = helper.RandString(7)
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})
		JustAfterEach(func() {
			helper.DeleteProject(project)
			helper.DeleteDir(context)
			helper.Chdir(originalDir)
		})
		It("should be able to list services in a given app and project combination", func() {
			// create a component by copying the example
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "create", "nodejs", "--app", app, "--project", project)

			// create a service from within a component directory
			helper.CmdShouldPass("odo", "service", "create", "dh-prometheus-apb", "--plan", "ephemeral",
				"--app", app, "--project", project,
			)
			ocArgs := []string{"get", "serviceinstance", "-o", "name", "-n", project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "dh-prometheus-apb")
			})

			// Listing the services should work as expected from within the component directory.
			// This means, it should not require --app or --project flags
			stdOut := helper.CmdShouldPass("odo", "service", "list")
			Expect(stdOut).To(ContainSubstring("dh-prometheus-apb"))

			// cd to a non-component directory and list services
			helper.Chdir(originalDir)
			stdOut = helper.CmdShouldPass("odo", "service", "list", "--app", app, "--project", project)
			Expect(stdOut).To(ContainSubstring("dh-prometheus-apb"))
		})

		It("should be able to create, list and delete services without a context and using --app and --project flags instaed", func() {
			// create a service using only app and project flags
			// we do Chdir first because originalDir doesn't have a context
			helper.Chdir(originalDir)

			// create the service
			helper.CmdShouldPass("odo", "service", "create", "dh-postgresql-apb", "--plan", "dev",
				"-p", "postgresql_user=luke", "-p", "postgresql_password=secret",
				"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6",
				"--app", app, "--project", project)

			ocArgs := []string{"get", "serviceinstance", "-o", "name", "-n", project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "dh-postgresql-apb")
			})

			// list the service using app and project flags
			stdOut := helper.CmdShouldPass("odo", "service", "list", "--app", app, "--project", project)
			Expect(stdOut).To(ContainSubstring("dh-postgresql-apb"))

			// delete the service using app and project flags
			helper.CmdShouldPass("odo", "service", "delete", "-f", "dh-postgresql-apb", "--app", app, "--project", project)
		})
	})
})
